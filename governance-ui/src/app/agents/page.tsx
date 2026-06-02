import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { Badge } from "@/components/ui/badge"
import { Bot, ShieldCheck } from "lucide-react"

export default function AgentDirectory() {
  const agents = [
    { did: "did:key:z6MkhaXgBZDvotDkL5257faiztiuC2ZX...", status: "Active", reputation: 98, lastActive: "2 mins ago" },
    { did: "did:key:z6MkqGz5qBZDvotDkL5257faiztiuC2ZY...", status: "Inactive", reputation: 85, lastActive: "1 day ago" },
    { did: "did:key:z6MkABC5qBZDvotDkL5257faiztiuC2ZZ...", status: "Active", reputation: 92, lastActive: "5 mins ago" },
    { did: "did:key:z6MkDEF5qBZDvotDkL5257faiztiuC2XW...", status: "Suspended", reputation: 45, lastActive: "1 week ago" },
  ]

  return (
    <div className="flex flex-col gap-8">
      <div>
        <h1 className="text-3xl font-bold tracking-tight">Agent Directory</h1>
        <p className="text-muted-foreground mt-2">
          Manage registered agents and monitor their reputation scores.
        </p>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Registered Agents</CardTitle>
          <CardDescription>
            A list of all agents authenticated in the Acumius network.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="w-[300px]">Agent DID</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Reputation Score</TableHead>
                <TableHead className="text-right">Last Active</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {agents.map((agent) => (
                <TableRow key={agent.did}>
                  <TableCell className="font-mono text-xs">{agent.did}</TableCell>
                  <TableCell>
                    <Badge variant={agent.status === 'Active' ? 'default' : agent.status === 'Suspended' ? 'destructive' : 'secondary'}>
                      {agent.status}
                    </Badge>
                  </TableCell>
                  <TableCell>
                    <div className="flex items-center gap-2">
                      <div className="flex-1 h-2 bg-secondary rounded-full overflow-hidden max-w-[100px]">
                        <div 
                          className={`h-full ${agent.reputation > 90 ? 'bg-emerald-500' : agent.reputation > 70 ? 'bg-yellow-500' : 'bg-destructive'}`} 
                          style={{ width: `${agent.reputation}%` }} 
                        />
                      </div>
                      <span className="text-sm font-medium">{agent.reputation}</span>
                    </div>
                  </TableCell>
                  <TableCell className="text-right text-muted-foreground">{agent.lastActive}</TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </div>
  )
}
