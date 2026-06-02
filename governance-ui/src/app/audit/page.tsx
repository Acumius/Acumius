import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { Badge } from "@/components/ui/badge"
import { Input } from "@/components/ui/input"
import { Search } from "lucide-react"

export default function AuditLog() {
  const logs = [
    { id: "LOG-901", action: "READ_MEMORY", agent: "did:key:z6MkhaXgBZD...", resource: "namespace:default", status: "ALLOWED", timestamp: "2026-06-03 14:32:01" },
    { id: "LOG-902", action: "WRITE_MEMORY", agent: "did:key:z6MkqGz5qBZD...", resource: "namespace:finance", status: "DENIED", timestamp: "2026-06-03 14:30:45" },
    { id: "LOG-903", action: "DELETE_MEMORY", agent: "did:key:z6MkABC5qBZD...", resource: "namespace:default", status: "ALLOWED", timestamp: "2026-06-03 14:15:22" },
  ]

  return (
    <div className="flex flex-col gap-8">
      <div>
        <h1 className="text-3xl font-bold tracking-tight">Audit Log</h1>
        <p className="text-muted-foreground mt-2">
          Immutable ledger of all policy engine decisions and agent actions.
        </p>
      </div>

      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle>System Activity</CardTitle>
              <CardDescription>Recent events recorded by the Audit logger.</CardDescription>
            </div>
            <div className="flex items-center space-x-2">
              <Input placeholder="Filter by Agent DID or Action..." className="w-[300px]" />
            </div>
          </div>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Timestamp</TableHead>
                <TableHead>Action</TableHead>
                <TableHead>Agent</TableHead>
                <TableHead>Resource</TableHead>
                <TableHead>Status</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {logs.map((log) => (
                <TableRow key={log.id}>
                  <TableCell className="text-muted-foreground text-sm">{log.timestamp}</TableCell>
                  <TableCell className="font-medium">{log.action}</TableCell>
                  <TableCell className="font-mono text-xs">{log.agent}</TableCell>
                  <TableCell>{log.resource}</TableCell>
                  <TableCell>
                    <Badge variant={log.status === 'ALLOWED' ? 'default' : 'destructive'}>
                      {log.status}
                    </Badge>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </div>
  )
}
