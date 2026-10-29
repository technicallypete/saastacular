import { Alert, AlertDescription } from "@/components/ui/alert";

type StatusMessageProps = {
  // Fixed text chosen by the page from a closed map; never a value read from the request.
  message: string | undefined;
  className?: string;
};

// A page-level outcome message in a polite status region: the kit Alert, whose own role is replaced by status (an
// outcome notice is polite, never assertive). Renders nothing without a message. className carries layout only.
export function StatusMessage({ message, className }: StatusMessageProps) {
  if (message === undefined) return null;
  return (
    <Alert role="status" data-testid="status-message" className={className}>
      <AlertDescription>{message}</AlertDescription>
    </Alert>
  );
}
