import { useRouteError } from "react-router-dom";
import { AlertCircle } from "lucide-react";
import { Alert, AlertTitle, AlertDescription } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";

export function GlobalErrorBoundary() {
  const error = useRouteError() as any;

  return (
    <div className="flex h-screen w-full items-center justify-center p-4 bg-background">
      <Alert variant="destructive" className="max-w-md shadow-lg">
        <AlertCircle className="h-4 w-4" />
        <AlertTitle className="font-bold">Unexpected Application Error!</AlertTitle>
        <AlertDescription className="mt-2 text-sm flex flex-col gap-4">
          <span className="font-mono bg-destructive/10 p-2 rounded block">
            {error?.message || error?.statusText || "An unknown error occurred while rendering the page."}
          </span>
          <Button variant="outline" size="sm" onClick={() => window.location.href = '/'}>
            Return to Dashboard
          </Button>
        </AlertDescription>
      </Alert>
    </div>
  );
}
