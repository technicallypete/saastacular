#!/usr/bin/env python3
"""Fixtures for check_sdlc.py in temporary directories; standard library unittest only.

Run: python3 -B -m unittest discover -s scripts -p 'test_check_sdlc.py' -v
Forbidden strings are assembled at run time so this file passes its own content scan.
"""
import contextlib
import hashlib
import io
import json
import os
from pathlib import Path
import shutil
import tempfile
import unittest

import check_sdlc as cs

SKILL = Path(__file__).resolve().parent.parent
HOME_PATH = "/" + "home" + "/alice/project"
REAL_UUID = "-".join(["1b4e28ba", "2fa1", "11d2", "883f", "0016d3cca427"])
SYNTHETIC = "00000000-0000-4000-8000-000000000001"
REMOTE = "git" + "lab.com/" + "group/repo"
KEY = "-" * 5 + "BEGIN RSA PRIV" + "ATE KEY" + "-" * 5


def run(argv):
    out = io.StringIO()
    with contextlib.redirect_stdout(out):
        code = cs.main([str(a) for a in argv])
    return code, json.loads(out.getvalue())


def write(path, text):
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(text)


def stub_companions(skills_root):
    write(skills_root / "claude-session-handoff" / "SKILL.md",
          "---\nname: claude-session-handoff\ndescription: Stub.\n---\n\nSee [ref](references/local-session-handoff.md).\n")
    write(skills_root / "claude-session-handoff" / "references" / "local-session-handoff.md", "# Reference\n")
    write(skills_root / "repository-supervisor" / "SKILL.md",
          "---\nname: repository-supervisor\ndescription: Stub.\n---\n\n"
          "Load [handoff](../claude-session-handoff/SKILL.md).\n")


def make_skills_root(tmp):
    root = Path(tmp) / "skills"
    shutil.copytree(SKILL, root / "sdlc-management",
                    ignore=shutil.ignore_patterns("__pycache__", "*.pyc"))
    stub_companions(root)
    return root


CODEX_CONFIG = ".codex/config.toml"
SPEC_CODEX_TOML = (
    "# Shared model defaults for this repository. Individual sessions may override them.\n"
    "model = \"gpt-6-luna\"\n"
    "model_reasoning_effort = \"medium\"\n"
    "\n"
    "[agents]\n"
    "default_subagent_model = \"gpt-6.1-sol\"\n"
    "default_subagent_reasoning_effort = \"high\"\n"
)


def adopt(repo, skills_root, release="v0", overrides=None, preserve=False):
    """Render every non-instance template into repo, install the bundle and add one Goal instance.

    Mirrors the documented procedure: the optional .codex/config.toml renders only when CODEX_CONFIG_TEMPLATE is
    true; with preserve=True an existing destination is kept and the rendering goes to <path>.sdlc-proposed.
    """
    skill = skills_root / "sdlc-management"
    values = cs.load_parameters(skill)
    values.update(overrides or {})
    values["RELEASE"] = release
    templates = skill / "templates"
    for tmpl in cs.walk_files(templates):
        rel = tmpl.relative_to(templates).as_posix()
        if rel.endswith(".tmpl") and tmpl.name not in cs.INSTANCE_NAMES:
            dest = cs.destination(rel, release)
            if dest == CODEX_CONFIG and values.get("CODEX_CONFIG_TEMPLATE") is not True:
                continue
            target = repo / dest
            if preserve and target.exists():
                target = repo / (dest + ".sdlc-proposed")
            write(target, cs.render(tmpl.read_text(), values))
    for sub in cs.RELEASE_DIRS:
        (repo / "management" / release / sub).mkdir(parents=True, exist_ok=True)
    # Procedure step 5: append the template's ignore lines that are missing; existing content is preserved.
    lines = [l for l in (templates / "root" / "gitignore.lines").read_text().splitlines() if l and l[0] != "#"]
    ignore = repo / ".gitignore"
    existing = ignore.read_text() if ignore.exists() else ""
    missing = [l for l in lines if l not in set(existing.splitlines())]
    if missing:
        prefix = existing if not existing or existing.endswith("\n") else existing + "\n"
        write(ignore, prefix + "\n".join(missing) + "\n")
    goal = cs.render((templates / "management/release/goals/GOAL.md.tmpl").read_text(), values)
    write(repo / "management" / release / "goals" / "G001-example.md",
          goal.replace("# G### — `<Goal title>`", "# G001 — Example goal"))
    shutil.copytree(skills_root, repo / ".agents" / "skills" / cs.BUNDLE_NAME)
    return values


PARENT_SKILL = ("---\nname: restruct-sdlc\ndescription: Stub parent.\n---\n\n"
                "Read [management](sdlc-management/SKILL.md).\n")


def make_bundle(tmp, skills_root):
    bundle = Path(tmp) / "bundle"
    parent = bundle / cs.BUNDLE_NAME
    shutil.copytree(skills_root, parent)
    write(parent / "SKILL.md", PARENT_SKILL)
    shutil.copy(SKILL / "LICENSE", parent / "LICENSE")
    write(parent / "agents" / "openai.yaml", "interface:\n  display_name: \"Stub\"\n")
    write(bundle / "README.md", "# Bundle\n\nStart with [the skill](restruct-sdlc/sdlc-management/SKILL.md).\n")
    shutil.copy(SKILL / "LICENSE", bundle / "LICENSE")
    rows = []
    for path in cs.walk_files(bundle):
        rel = path.relative_to(bundle).as_posix()
        rows.append(f"{hashlib.sha256(path.read_bytes()).hexdigest()}\tnew\t{rel}")
    write(bundle / "MANIFEST.tsv", "\n".join(rows) + "\n")
    return bundle


def detail(result, check):
    return [f["detail"] for f in result["findings"] if f["check"] == check]


def assert_only_p2_baseline(case, code, result, checks):
    """D5 oracle (Plan r4): the preserved 220-line sdlc-management/SKILL.md is a documented baseline P2 failure.

    Exit 1, P2 the only failing check, every other check passing, and exactly one finding naming that file. This is
    recorded as a baseline failure, never as a pass; MAX_SKILL_LINES stays 200 and the child is not trimmed.
    """
    case.assertEqual(code, 1, result["findings"])
    case.assertEqual(result["checks"], {c: ("fail" if c == "P2" else "pass") for c in sorted(checks)})
    case.assertEqual(len(result["findings"]), 1, result["findings"])
    finding = result["findings"][0]
    case.assertEqual((finding["check"], finding["detail"]), ("P2", "220 lines exceeds 200"))
    case.assertEqual(Path(finding["path"]).parts[-2:], ("sdlc-management", "SKILL.md"))


PACKAGE_CHECKS = {"P1", "P2", "P3", "P3t", "P4", "P5", "P6"}


class PackageChecks(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory()
        self.root = make_skills_root(self.tmp.name)
        self.skill = self.root / "sdlc-management"

    def tearDown(self):
        self.tmp.cleanup()

    def package(self, *extra):
        return run(["package", self.skill, "--skills-root", self.root, *extra])

    def baseline(self, *extra):
        assert_only_p2_baseline(self, *self.package(*extra), PACKAGE_CHECKS)

    def test_real_skill_passes_all_checks(self):
        self.baseline()

    def test_p1_name_must_match_directory(self):
        path = self.skill / "SKILL.md"
        path.write_text(path.read_text().replace("name: sdlc-management", "name: other-name", 1))
        code, result = self.package()
        self.assertEqual(code, 1)
        self.assertTrue(detail(result, "P1"))

    def test_p1_missing_frontmatter_and_description(self):
        path = self.skill / "SKILL.md"
        text = path.read_text()
        path.write_text(text.replace("description:", "summary:", 1))
        self.assertTrue(detail(self.package()[1], "P1"))
        path.write_text(text.replace("---\n", "", 1))
        self.assertIn("frontmatter not delimited by --- lines", detail(self.package()[1], "P1"))

    def test_p2_line_limit(self):
        path = self.skill / "SKILL.md"
        path.write_text(path.read_text() + "\n" * 200)
        self.assertTrue(detail(self.package()[1], "P2"))

    def test_p3_shipped_missing_target_fails(self):
        write(self.skill / "references" / "budgets.md",
              (self.skill / "references" / "budgets.md").read_text() + "\n[gone](missing.md)\n")
        self.assertTrue(any("does not exist" in d for d in detail(self.package()[1], "P3")))

    def test_p3_shipped_escape_and_absolute_fail(self):
        doc = self.skill / "references" / "adoption.md"
        doc.write_text(doc.read_text() + "\n[up](../../../outside.md)\n[abs](/etc/passwd)\n")
        details = detail(self.package()[1], "P3")
        self.assertTrue(any("escapes" in d for d in details), details)
        self.assertTrue(any("absolute" in d for d in details), details)

    def test_p3_sibling_links_resolve_only_against_declared_root(self):
        doc = self.skill / "references" / "coordination.md"
        doc.write_text(doc.read_text() + "\n[x](../../unknown-skill/SKILL.md)\n")
        self.assertTrue(any("unknown skill" in d for d in detail(self.package()[1], "P3")))
        doc.write_text(doc.read_text().replace("../../unknown-skill/SKILL.md", "../../repository-supervisor/NOPE.md"))
        self.assertTrue(any("does not exist" in d for d in detail(self.package()[1], "P3")))
        shutil.rmtree(self.root / "repository-supervisor")
        self.assertTrue(detail(self.package()[1], "P3"))

    def test_p3_links_inside_code_are_ignored(self):
        doc = self.skill / "references" / "budgets.md"
        doc.write_text(doc.read_text() + "\n`[a](missing.md)`\n\n```\n[b](missing.md)\n```\n")
        self.baseline()

    def test_p3t_rendered_template_in_documented_layout_passes(self):
        tmpl = self.skill / "templates" / "root" / "AGENTS.md.tmpl"
        tmpl.write_text(tmpl.read_text() + "\n[progress](management/{{RELEASE}}/PROGRESS.md)\n")
        self.baseline()

    def test_p3t_undeclared_and_malformed_parameter_fail(self):
        tmpl = self.skill / "templates" / "docs" / "RISKS.md.tmpl"
        tmpl.write_text(tmpl.read_text() + "\n{{UNDECLARED_NAME}} {{ lower }} {{OPEN\n")
        details = detail(self.package()[1], "P3t")
        self.assertTrue(any("undeclared parameter 'UNDECLARED_NAME'" in d for d in details), details)
        self.assertTrue(any("malformed placeholder" in d for d in details), details)
        self.assertTrue(any("unbalanced" in d for d in details), details)

    def test_p3t_missing_target_fails(self):
        tmpl = self.skill / "templates" / "root" / "CLAUDE.md.tmpl"
        tmpl.write_text(tmpl.read_text() + "\n[x](docs/NOT-CREATED.md)\n"
                        "[y](.agents/skills/restruct-sdlc/sdlc-management/references/nope.md)\n")
        details = detail(self.package()[1], "P3t")
        self.assertEqual(len(details), 2, details)

    def test_p3t_escaping_target_and_scheme_fail(self):
        tmpl = self.skill / "templates" / "management" / "release" / "PROGRESS.md.tmpl"
        tmpl.write_text(tmpl.read_text() + "\n[x](../../../outside.md)\n[y](https://example.org/)\n")
        details = detail(self.package()[1], "P3t")
        self.assertTrue(any("escapes" in d for d in details), details)
        self.assertTrue(any("must be relative" in d for d in details), details)

    def test_p4_unreachable_reference_fails(self):
        write(self.skill / "references" / "orphan.md", "# Orphan\n")
        self.assertTrue(detail(self.package()[1], "P4"))

    def test_p5_symlink_bytecode_and_git_fail(self):
        os.symlink("SKILL.md", self.skill / "link.md")
        write(self.skill / "scripts" / "__pycache__" / "x.pyc", "")
        (self.skill / ".git").mkdir()
        details = detail(self.package()[1], "P5")
        for label in ("symlink", "bytecode", "Git metadata"):
            self.assertIn(label, details)

    def test_p6_forbidden_content_and_literals(self):
        doc = self.skill / "references" / "budgets.md"
        base = doc.read_text()
        for text, label in ((HOME_PATH, "home-path"), (REMOTE, "hosted-remote"), (KEY, "private-key"),
                            (REAL_UUID, "non-synthetic full UUID")):
            doc.write_text(base + "\n" + text + "\n")
            self.assertTrue(any(label in d for d in detail(self.package()[1], "P6")), label)
        doc.write_text(base + "\n" + SYNTHETIC + "\n")
        self.baseline()
        literals = Path(self.tmp.name) / "literals.txt"
        word = "Acme" + "Corp"
        literals.write_text("# private\n" + word + "\n")
        doc.write_text(base + "\nWorks for " + word.lower() + " today.\n")
        self.assertIn("private literal entry #2", detail(self.package("--literals", literals)[1], "P6"))
        doc.write_text(base + "\nnotacmecorpx\n")
        self.baseline("--literals", literals)

    def test_usage_errors_exit_2(self):
        missing = Path(self.tmp.name) / "none"
        for argv in (["package", missing], ["nope"], ["adoption", self.tmp.name], ["package", self.skill,
                     "--literals", missing], ["adoption", self.tmp.name, "--release", "../x"]):
            with contextlib.redirect_stderr(io.StringIO()), self.assertRaises(SystemExit) as ctx:
                cs.main([str(a) for a in argv])
            self.assertEqual(ctx.exception.code, 2, argv)


class BundleChecks(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory()
        self.bundle = make_bundle(self.tmp.name, make_skills_root(self.tmp.name))

    def tearDown(self):
        self.tmp.cleanup()

    def test_bundle_passes(self):
        assert_only_p2_baseline(self, *run(["package", self.bundle]), PACKAGE_CHECKS | {"P7"})

    def test_bundle_parent_p1_p2(self):
        """Bundle mode checks the parent entrypoint itself: P1 name/description and the 200-line P2 limit."""
        parent = self.bundle / cs.BUNDLE_NAME / "SKILL.md"
        parent.write_text(PARENT_SKILL.replace("name: restruct-sdlc", "name: other-parent") + "\n" * 200)
        result = run(["package", self.bundle])[1]
        parent_findings = {(f["check"], f["detail"]) for f in result["findings"] if f["path"] == str(parent)}
        self.assertTrue(any(c == "P1" and "must equal the directory name" in d for c, d in parent_findings),
                        result["findings"])
        self.assertTrue(any(c == "P2" and d.endswith("exceeds 200") for c, d in parent_findings), result["findings"])

    def test_p7_extra_file_hash_mismatch_and_bad_origin(self):
        write(self.bundle / cs.BUNDLE_NAME / "repository-supervisor" / "extra.md", "x\n")
        manifest = self.bundle / "MANIFEST.tsv"
        lines = manifest.read_text().splitlines()
        lines[0] = "0" * 64 + "\tnew\t" + lines[0].split("\t")[2]
        lines[1] = lines[1].replace("\tnew\t", "\tcopied\t")
        manifest.write_text("\n".join(lines) + "\n")
        details = detail(run(["package", self.bundle])[1], "P7")
        self.assertTrue(any("not listed" in d for d in details), details)
        self.assertTrue(any("hash mismatch" in d for d in details), details)
        self.assertTrue(any("invalid origin" in d for d in details), details)

    def test_p7_root_and_skill_set(self):
        write(self.bundle / "NOTES.md", "x\n")
        shutil.rmtree(self.bundle / cs.BUNDLE_NAME / "repository-supervisor")
        details = detail(run(["package", self.bundle])[1], "P7")
        self.assertTrue(any("root entries" in d for d in details), details)
        self.assertTrue(any("bundle entries [" in d for d in details), details)

    def test_bundle_readme_link_must_stay_inside(self):
        readme = self.bundle / "README.md"
        readme.write_text(readme.read_text() + "\n[x](../outside.md)\n")
        self.assertTrue(detail(run(["package", self.bundle])[1], "P3"))


class AdoptionChecks(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory()
        self.root = make_skills_root(self.tmp.name)
        self.repo = Path(self.tmp.name) / "repo"
        self.repo.mkdir()
        adopt(self.repo, self.root)

    def tearDown(self):
        self.tmp.cleanup()

    def adoption(self, release="v0", *extra):
        return run(["adoption", self.repo, "--release", release, *extra])

    def test_rendered_adoption_passes(self):
        code, result = self.adoption()
        self.assertEqual(code, 0, result["findings"])
        self.assertEqual(set(result["checks"]), {"A1", "A2", "A3", "A4", "A5", "A6"})

    def test_other_release_name(self):
        repo = Path(self.tmp.name) / "repo-r1"
        repo.mkdir()
        adopt(repo, self.root, release="r1", overrides={"INTEGRATION_BRANCH": "develop", "RELEASE_BRANCH": "release"})
        code, result = run(["adoption", repo, "--release", "r1"])
        self.assertEqual(code, 0, result["findings"])
        self.assertEqual(run(["adoption", repo, "--release", "v0"])[0], 1)

    def test_a1_missing_paths(self):
        (self.repo / "management" / "v0" / "BACKLOG.md").unlink()
        shutil.rmtree(self.repo / "management" / "v0" / "archive")
        self.assertEqual(len(detail(self.adoption()[1], "A1")), 2)

    def test_a2_rendered_missing_target_fails(self):
        (self.repo / "docs" / "RISKS.md").unlink()
        self.assertTrue(any("does not exist" in d for d in detail(self.adoption()[1], "A2")))

    def test_a2_rendered_escape_fails(self):
        agents = self.repo / "AGENTS.md"
        agents.write_text(agents.read_text() + "\n[x](../../outside.md)\n")
        self.assertTrue(any("escapes" in d for d in detail(self.adoption()[1], "A2")))

    def test_a2_installed_skill_links_checked(self):
        shutil.rmtree(self.repo / ".agents" / "skills" / cs.BUNDLE_NAME / "claude-session-handoff" / "references")
        details = detail(self.adoption()[1], "A2")
        self.assertTrue(any(d.startswith("installed skill:") for d in details), details)

    def test_a3_unresolved_placeholder_fails(self):
        progress = self.repo / "management" / "v0" / "PROGRESS.md"
        progress.write_text(progress.read_text() + "\n{{PROJECT_NAME}}\n")
        self.assertTrue(detail(self.adoption()[1], "A3"))

    def test_a4_goal_fields_in_order(self):
        goal = self.repo / "management" / "v0" / "goals" / "G001-example.md"
        text = goal.read_text()
        goal.write_text(text.replace("## Objective", "## Purpose"))
        self.assertIn("field 'Objective' missing or out of order", detail(self.adoption()[1], "A4"))
        swapped = text.replace("## In Scope", "## TMP").replace("## Out of Scope", "## In Scope")
        goal.write_text(swapped.replace("## TMP", "## Out of Scope"))
        self.assertTrue(detail(self.adoption()[1], "A4"))
        goal.write_text(text.replace("# G001 — Example goal", "# Example goal"))
        self.assertTrue(detail(self.adoption()[1], "A4"))

    def test_a5_ignore_lines(self):
        (self.repo / ".gitignore").write_text("node_modules/\n")
        self.assertEqual(len(detail(self.adoption()[1], "A5")), 5)

    def test_a5_each_required_ignore_line_is_enforced(self):
        """Every template exclusion, including /.claude/settings.local.json and /.local/, is required by A5."""
        full = (self.repo / ".gitignore").read_text().splitlines()
        self.assertIn("/.claude/settings.local.json", cs.IGNORE_LINES)
        self.assertIn("/.local/", cs.IGNORE_LINES)
        for line in cs.IGNORE_LINES:
            (self.repo / ".gitignore").write_text("\n".join(l for l in full if l != line) + "\n")
            self.assertEqual(detail(self.adoption()[1], "A5"), [f"missing ignore line {line!r}"], line)
        (self.repo / ".gitignore").write_text("\n".join(full) + "\n")
        self.assertEqual(self.adoption()[0], 0)

    def test_bootstrap_appends_missing_ignore_lines(self):
        """Procedure step 5: existing .gitignore content survives; each required line is added exactly once."""
        repo = Path(self.tmp.name) / "ignore-existing"
        original = "node_modules/\n/.local/\n# project rule\ndist/"
        write(repo / ".gitignore", original)
        adopt(repo, self.root)
        text = (repo / ".gitignore").read_text()
        self.assertTrue(text.startswith(original + "\n"), text)
        lines = text.splitlines()
        for line in cs.IGNORE_LINES:
            self.assertEqual(lines.count(line), 1, line)
        self.assertEqual(lines.count("node_modules/"), 1)
        shutil.rmtree(repo / ".agents")  # the helper installs the bundle by copy; repeat only the rendering steps
        adopt(repo, self.root)
        self.assertEqual((repo / ".gitignore").read_text(), text, "a repeated adoption adds nothing")
        self.assertEqual(detail(run(["adoption", repo, "--release", "v0"])[1], "A5"), [])

    def test_a6_literal_scan(self):
        literals = Path(self.tmp.name) / "literals.txt"
        literals.write_text("Example Project\n")
        self.assertTrue(detail(self.adoption("v0", "--literals", literals)[1], "A6"))

    def test_existing_application_content_survives_adoption(self):
        """C1: unrelated app delimiters, a public source URL and a real UUID are legitimate target content."""
        app = {"app/templates/invoice.html": "<p>{{ customer.name }}</p>{{#items}}{{sku}}{{/items}}\n",
               "README.md": "Source: https://" + "githu" + "b.com/example/project\n",
               "docs/notes.md": "Tenant " + REAL_UUID + " and " + HOME_PATH + " are project facts.\n",
               "management/v0/results/G001-notes.md": "Record " + REAL_UUID + " from https://" + REMOTE + "\n"}
        for rel, text in app.items():
            write(self.repo / rel, text)
        before = {rel: (self.repo / rel).read_bytes() for rel in app}
        code, result = self.adoption()
        self.assertEqual(code, 0, result["findings"])
        self.assertEqual(before, {rel: (self.repo / rel).read_bytes() for rel in app})

    def test_adopted_artifact_placeholders_still_fail(self):
        write(self.repo / "app" / "page.html", "{{ title }}\n")
        for rel in ("AGENTS.md.sdlc-proposed", "management/v0/specs/G001-example-spec.md", "docs/RISKS.md"):
            write(self.repo / rel, "{{UNDECLARED_X}}\n")
        paths = {Path(f["path"]).relative_to(self.repo).as_posix() for f in self.adoption()[1]["findings"]
                 if f["check"] == "A3"}
        self.assertEqual(paths, {"AGENTS.md.sdlc-proposed", "management/v0/specs/G001-example-spec.md",
                                 "docs/RISKS.md"})

    def test_a6_source_origin_and_private_state_scope(self):
        literals = Path(self.tmp.name) / "origin.txt"
        word = "Origin" + "Corp"
        literals.write_text(word + "\n")
        write(self.repo / "app" / "config.txt", "owner=" + word.lower() + "\n")
        self.assertTrue(any("private literal" in d for d in detail(self.adoption("v0", "--literals", literals)[1], "A6")))
        (self.repo / "app" / "config.txt").unlink()
        write(self.repo / "app" / "id_rsa", KEY + "\n")
        self.assertTrue(any("private-key" in d for d in detail(self.adoption()[1], "A6")))
        (self.repo / "app" / "id_rsa").unlink()
        skill_doc = self.repo / ".agents" / "skills" / cs.BUNDLE_NAME / "sdlc-management" / "references" / "budgets.md"
        for planted, label in ((HOME_PATH, "home-path"), (REAL_UUID, "non-synthetic full UUID"),
                               ("https://" + REMOTE, "hosted-remote")):
            original = skill_doc.read_text()
            skill_doc.write_text(original + planted + "\n")
            self.assertTrue(any(label in d for d in detail(self.adoption()[1], "A6")), label)
            skill_doc.write_text(original)
        self.assertEqual(self.adoption()[0], 0)

    def test_private_worker_checkouts_are_not_adopted_content(self):
        worker = self.repo / ".claude" / "worktrees" / "g002-agent-x-impl"
        write(worker / "management" / "notes.md", "{{PROJECT_NAME}} " + HOME_PATH + "\n")
        write(self.repo / ".codex" / "worktrees" / "g002-agent-y-check" / "x.md", "{{RELEASE}}\n")
        self.assertEqual(self.adoption()[0], 0)

    def test_bootstrap_creates_missing_files(self):
        """AC3(a): explicit adoption into an empty repository creates every layout file with the selected values."""
        repo = Path(self.tmp.name) / "fresh"
        repo.mkdir()
        selected = {"CODEX_CONFIG_TEMPLATE": True, "PROJECT_NAME": "Bootstrap Demo", "IMPLEMENTER_AGENT": "Impl Agent X",
                    "CODEX_MAIN_MODEL": "main-model-x", "CODEX_SUBAGENT_EFFORT": "xhigh"}
        adopt(repo, self.root, overrides=selected)
        templates = self.root / "sdlc-management" / "templates"
        layout = cs.adoption_layout(templates, "v0")
        missing = sorted(rel for rel in layout if not (repo / rel).exists())
        self.assertEqual(missing, [])
        created = {p.relative_to(repo).as_posix() for p in cs.walk_files(repo)
                   if not p.relative_to(repo).as_posix().startswith(".agents/")}
        self.assertEqual(created - layout, {"management/v0/goals/G001-example.md"})
        self.assertEqual([p for p in created if p.endswith(".sdlc-proposed")], [])
        text = {rel: (repo / rel).read_text() for rel in ("AGENTS.md", "CLAUDE.md", CODEX_CONFIG,
                                                           "management/v0/ROADMAP.md")}
        self.assertIn("Bootstrap Demo", text["AGENTS.md"])
        self.assertIn("Impl Agent X", text["CLAUDE.md"])
        self.assertIn('model = "main-model-x"', text[CODEX_CONFIG])
        self.assertIn('default_subagent_reasoning_effort = "xhigh"', text[CODEX_CONFIG])
        self.assertIn("Bootstrap Demo", text["management/v0/ROADMAP.md"])
        # AC4(iv): each owner workflow rule reaches adopters at its actual template location.
        flat = {rel: " ".join(body.split()) for rel, body in text.items()}
        handoff = " ".join((repo / "management" / "SUPERVISOR-HANDOFF.md").read_text().split())
        self.assertIn("do not poll coordinator sessions or ping the session-handoff watcher unless the owner "
                      "specifically asks", flat["CLAUDE.md"])
        self.assertIn("send needed coordinator actions immediately using nonblocking delivery and continue without "
                      "waiting for a response; the watcher delivers acknowledgment asynchronously", flat["CLAUDE.md"])
        self.assertIn("## Apply workflow updates to live repository guidance", text["AGENTS.md"])
        self.assertIn("This applies to every adopting repository.", flat["AGENTS.md"])
        self.assertIn("the supervisor does not poll sessions or ping the watcher unless the owner specifically asks",
                      flat["AGENTS.md"])
        self.assertIn("Send needed coordinator decisions or requests immediately using nonblocking delivery.",
                      flat["AGENTS.md"])
        self.assertIn("using nonblocking delivery. Continue supervisor work without waiting for a response", handoff)
        code, result = run(["adoption", repo, "--release", "v0"])
        self.assertEqual(code, 0, result["findings"])

    def test_bootstrap_preserves_existing_root_files(self):
        """AC3(b)/AC10: existing root governance and Codex config stay byte-identical; drafts are proposed."""
        repo = Path(self.tmp.name) / "existing"
        originals = {"AGENTS.md": "# Project agents\n\nLocal rules.\n", "CLAUDE.md": "# Project Claude\n\nLocal.\n",
                     CODEX_CONFIG: "model = \"local-choice\"\n"}
        for rel, body in originals.items():
            write(repo / rel, body)
        before = {rel: hashlib.sha256((repo / rel).read_bytes()).hexdigest() for rel in originals}
        values = adopt(repo, self.root, overrides={"CODEX_CONFIG_TEMPLATE": True}, preserve=True)
        self.assertEqual(before, {rel: hashlib.sha256((repo / rel).read_bytes()).hexdigest() for rel in originals})
        templates = self.root / "sdlc-management" / "templates"
        sources = {"AGENTS.md": "root/AGENTS.md.tmpl", "CLAUDE.md": "root/CLAUDE.md.tmpl",
                   CODEX_CONFIG: "root/.codex/config.toml.tmpl"}
        for rel, tmpl in sources.items():
            proposal = repo / (rel + ".sdlc-proposed")
            self.assertEqual(proposal.read_text(), cs.render((templates / tmpl).read_text(), values), rel)
        layout = cs.adoption_layout(templates, "v0")
        self.assertEqual(sorted(rel for rel in layout if not (repo / rel).exists()), [])
        # AC10: rendered governance names the shared .local root and existing-task retention; the installed adoption
        # reference it links carries the supported isolated-worker fallback.
        for rel in ("AGENTS.md", "CLAUDE.md"):
            body = " ".join((repo / (rel + ".sdlc-proposed")).read_text().split())
            self.assertIn("Use the main integration checkout's ignored `.local/{evidence,secrets,tmp}/` for new "
                          "task-private files, shared across worker worktrees.", body, rel)
            self.assertIn("Existing tasks keep current paths through acceptance cleanup", body, rel)
            self.assertIn("references/adoption.md#repository-local-private-files", body, rel)
        reference = " ".join((repo / ".agents/skills" / cs.BUNDLE_NAME / "sdlc-management/references/adoption.md")
                             .read_text().split())
        self.assertIn("Supported isolated-worker fallback: a worker that cannot reach the main checkout's `.local/` "
                      "uses the documented home-state path recorded in its dispatch", reference)
        self.assertIn("never creates a per-worker `.local/`", reference)
        self.assertIn("/.local/", (templates / "root" / "gitignore.lines").read_text().splitlines())
        code, result = run(["adoption", repo, "--release", "v0"])
        self.assertEqual(code, 0, result["findings"])

    def test_codex_config_template_exact_default(self):
        """AC3(c): defaults render byte-equal to the Spec block; the file is opt-in; P3t layout includes it."""
        templates = self.root / "sdlc-management" / "templates"
        values = cs.load_parameters(self.root / "sdlc-management")
        self.assertIs(values["CODEX_CONFIG_TEMPLATE"], False)
        rendered = cs.render((templates / "root" / ".codex" / "config.toml.tmpl").read_text(), values)
        self.assertEqual(rendered, SPEC_CODEX_TOML)
        self.assertFalse((self.repo / CODEX_CONFIG).exists())
        self.assertIn(CODEX_CONFIG, cs.adoption_layout(templates, "v0"))

    def test_ignore_lines_constant_matches_template(self):
        lines = (SKILL / "templates" / "root" / "gitignore.lines").read_text().splitlines()
        self.assertEqual(tuple(l for l in lines if l and l[0] != "#"), cs.IGNORE_LINES)
        self.assertEqual(len(cs.IGNORE_LINES), 5)


if __name__ == "__main__":
    unittest.main()
