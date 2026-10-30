/* Opens an architecture diagram or a gallery screenshot in a native modal dialog. Same-origin, no libraries, no network, no storage.
   A diagram is shown as a copy of the page's own SVG with unique ids; a screenshot as its own image, from the same file. */
(function () {
  "use strict";
  var dialog = document.getElementById("diagram-dialog");
  var title = document.getElementById("diagram-dialog-title");
  var body = document.getElementById("diagram-dialog-body");
  if (!dialog || !title || !body || typeof dialog.showModal !== "function") return;
  var closeButton = dialog.querySelector(".diagram-close");
  var opener = null;
  var closing = false;

  function text(el) {
    return el.textContent.replace(/\s+/g, " ").trim();
  }

  function show(copy, label, from) {
    body.replaceChildren(copy);
    title.textContent = label;
    opener = from;
    dialog.showModal();
    body.scrollTop = 0;
  }

  function openDiagram(figure, from) {
    var svg = figure.querySelector("svg.a-svg");
    var caption = figure.querySelector("figcaption");
    if (!svg || !caption) return;
    var copy = svg.cloneNode(true);
    var renamed = {};
    Array.prototype.forEach.call(copy.querySelectorAll("[id]"), function (el) {
      renamed[el.id] = "ov-" + el.id;
      el.id = renamed[el.id];
    });
    var names = copy.getAttribute("aria-labelledby");
    if (names) {
      copy.setAttribute("aria-labelledby", names.split(/\s+/).map(function (id) { return renamed[id] || id; }).join(" "));
    }
    show(copy, text(caption), from);
  }

  function openShot(figure, from) {
    var img = figure.querySelector("img");
    var caption = figure.querySelector("figcaption");
    if (!img || !caption) return;
    var copy = img.cloneNode(false);
    copy.removeAttribute("loading");
    copy.style.maxWidth = copy.getAttribute("width") + "px";
    show(copy, text(caption), from);
  }

  Array.prototype.forEach.call(document.querySelectorAll("figure.diagram"), function (figure) {
    var button = figure.querySelector("[data-diagram-open]");
    if (!button) return;
    button.parentNode.hidden = false;
    figure.classList.add("diagram-zoomable");
    button.addEventListener("click", function () { openDiagram(figure, button); });
    figure.querySelector(".diagram-scroll").addEventListener("click", function (event) {
      if (event.target.closest && event.target.closest("svg")) openDiagram(figure, button);
    });
  });

  Array.prototype.forEach.call(document.querySelectorAll("figure.shot"), function (figure) {
    var img = figure.querySelector("img");
    if (!img) return;
    var button = document.createElement("button");
    button.type = "button";
    button.className = "shot-open";
    button.setAttribute("aria-haspopup", "dialog");
    button.setAttribute("aria-label", "Open full size: " + img.getAttribute("alt"));
    img.parentNode.insertBefore(button, img);
    button.appendChild(img);
    button.addEventListener("click", function () { openShot(figure, button); });
  });

  /* Close through the exit animation: the dialog stays modal until it has finished (no animation, so no wait, with reduced motion). */
  function finish() {
    dialog.classList.remove("closing");
    if (dialog.open) dialog.close();
  }

  function requestClose() {
    if (closing || !dialog.open) return;
    closing = true;
    dialog.classList.add("closing");
    Promise.all(dialog.getAnimations({ subtree: true }).map(function (a) { return a.finished; })).then(finish, finish);
  }

  closeButton.addEventListener("click", requestClose);
  dialog.addEventListener("click", function (event) { if (event.target === dialog) requestClose(); });
  dialog.addEventListener("cancel", function (event) {
    event.preventDefault();
    requestClose();
  });
  dialog.addEventListener("close", function () {
    closing = false;
    dialog.classList.remove("closing");
    body.replaceChildren();
    if (opener) opener.focus({ preventScroll: true });
    opener = null;
  });
})();
