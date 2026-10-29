"use client";

// Test-only client part of the token fixture page (page.tsx): the root surface's portaled kit components. Each one
// inserts its content into document.body, where only the document root's tenant values apply. Each kit dialog keeps the
// kit DialogContent's own corner Close (showCloseButton, its default; named "Close" by its sr-only text), as every app
// dialog does since the G069 owner kit-fidelity direction (2026-10-26); the dialog field is the bare kit Input.
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Popover, PopoverContent, PopoverDescription, PopoverHeader, PopoverTitle, PopoverTrigger } from "@/components/ui/popover";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";

const SELECT_ITEMS = [
  { label: "One", value: "one" },
  { label: "Two", value: "two" },
];

// open: the root dialog is open at mount (?open=root), for the first-insertion portal checks.
export function FixtureOverlays({ open }: { open: boolean }) {
  return (
    <div className="flex flex-wrap items-center gap-4">
      <Dialog>
        <DialogTrigger render={<Button variant="secondary" />} data-testid="dialog-trigger">
          Keyboard help
        </DialogTrigger>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Fixture dialog</DialogTitle>
            <DialogDescription>A kit dialog portalled into the document body.</DialogDescription>
          </DialogHeader>
          <Input aria-label="Dialog field" data-testid="dialog-input" />
          <DialogFooter>
            <Button data-testid="dialog-action">Confirm</Button>
            <DialogClose render={<Button variant="outline" />}>Done</DialogClose>
          </DialogFooter>
        </DialogContent>
      </Dialog>
      <Select items={SELECT_ITEMS} defaultValue="one">
        <SelectTrigger aria-label="Fixture select" data-testid="select-trigger">
          <SelectValue />
        </SelectTrigger>
        <SelectContent>
          {SELECT_ITEMS.map((item) => (
            <SelectItem key={item.value} value={item.value}>
              {item.label}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
      <Popover>
        <PopoverTrigger render={<Button variant="outline" />} data-testid="popover-trigger">
          Open popover
        </PopoverTrigger>
        <PopoverContent>
          <PopoverHeader>
            <PopoverTitle>Fixture popover</PopoverTitle>
            <PopoverDescription>A kit popover portalled into the document body.</PopoverDescription>
          </PopoverHeader>
        </PopoverContent>
      </Popover>
      <Button
        variant="outline"
        data-testid="toast-trigger"
        onClick={() => toast("Fixture toast", { description: "A kit toast.", duration: Infinity })}
      >
        Show toast
      </Button>
      {open ? (
        <Dialog defaultOpen>
          <DialogContent aria-modal="true">
            <DialogHeader>
              <DialogTitle>Root dialog</DialogTitle>
              <DialogDescription>Open at mount, inserted into the document body.</DialogDescription>
            </DialogHeader>
            <DialogFooter>
              <DialogClose render={<Button variant="outline" />}>Done</DialogClose>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      ) : null}
    </div>
  );
}
