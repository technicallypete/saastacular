"use client";

import * as React from "react";

// G019 (Plan r1 §1.6): fills the invitation code field from the link's fragment (#<token>) in the browser. A fragment is
// never sent to a server, so the code reaches this application only in the accept form's POST body. The fragment is then
// removed from the address bar (history.replaceState) so it is not left in the visible URL or later copied links.
const TOKEN = /^[A-Za-z0-9_-]{43}$/;

export function TokenPrefill({ fieldId }: { fieldId: string }) {
  React.useEffect(() => {
    const token = window.location.hash.slice(1);
    const input = document.getElementById(fieldId);
    if (!TOKEN.test(token) || !(input instanceof HTMLInputElement) || input.value !== "") return;
    input.value = token;
    window.history.replaceState(null, "", window.location.pathname + window.location.search);
  }, [fieldId]);
  return null;
}
