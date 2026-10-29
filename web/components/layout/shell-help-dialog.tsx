"use client";

import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle, DialogTrigger } from "@/components/ui/dialog";
import { BulletItem, BulletList } from "@/components/vestive/content";

const TIPS = [
  "Tab and Shift+Tab move between links and controls.",
  "The first Tab on a page reveals Skip to main content.",
  "Arrow keys choose between Light, Dark and System appearance.",
  "Escape or the Close button closes this dialog.",
];

// Generic keyboard help: a modal dialog that keeps focus inside while open and returns it to the
// trigger when closed. aria-modal is passed through to the dialog element. G069 (Plan r1 S7, S12): the kit dialog,
// whose overlay and content portal into document.body and take the document root's tenant values from their first
// frame; the corner Close is the kit DialogContent's own (showCloseButton, its default), named "Close" by its sr-only
// text.
export function ShellHelpDialog() {
  return (
    <Dialog>
      <DialogTrigger render={<Button size="sm" />}>Keyboard help</DialogTrigger>
      <DialogContent aria-modal="true">
        <DialogHeader>
          <DialogTitle>Keyboard help</DialogTitle>
          <DialogDescription>Everything here works with a keyboard alone.</DialogDescription>
        </DialogHeader>
        <BulletList>
          {TIPS.map((tip) => (
            <BulletItem key={tip}>{tip}</BulletItem>
          ))}
        </BulletList>
      </DialogContent>
    </Dialog>
  );
}
