import { Logs } from "lucide-react";
import React from "react";

import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";

export function ListBackendApiKeyAuditLogsCard() {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Logs className="w-4 h-4" />
          <span>Audit Logs</span>
        </CardTitle>
        <CardDescription>
          Events assiciated with this Backend API Key.
        </CardDescription>
      </CardHeader>
      <CardContent>
        {/* Content for listing backend API key audit logs will go here */}
        <p className="text-muted-foreground">No audit logs available yet.</p>
      </CardContent>
    </Card>
  );
}
