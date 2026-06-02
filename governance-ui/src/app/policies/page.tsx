import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Shield, Plus } from "lucide-react"

export default function PoliciesPage() {
  const policies = [
    { id: "P-101", name: "Default Deny", type: "System", status: "Active" },
    { id: "P-102", name: "Agent Read-Only", type: "Memory Access", status: "Active" },
    { id: "P-103", name: "Restricted Namespace", type: "Namespace", status: "Inactive" },
  ]

  return (
    <div className="flex flex-col gap-8">
      <div className="flex justify-between items-start">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Policies</h1>
          <p className="text-muted-foreground mt-2">
            Manage access control policies for agents and memory namespaces.
          </p>
        </div>
        <Button>
          <Plus className="mr-2 h-4 w-4" /> Create Policy
        </Button>
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        {policies.map((policy) => (
          <Card key={policy.id}>
            <CardHeader className="pb-2">
              <div className="flex items-center justify-between">
                <CardTitle className="text-lg">{policy.name}</CardTitle>
                <Shield className="h-4 w-4 text-muted-foreground" />
              </div>
              <CardDescription>{policy.type} Policy</CardDescription>
            </CardHeader>
            <CardContent>
              <div className="mt-4 flex justify-between items-center text-sm">
                <span className={`px-2 py-1 rounded-full text-xs font-medium ${policy.status === 'Active' ? 'bg-emerald-500/10 text-emerald-500' : 'bg-secondary text-muted-foreground'}`}>
                  {policy.status}
                </span>
                <Button variant="link" className="p-0">Edit Policy</Button>
              </div>
            </CardContent>
          </Card>
        ))}
      </div>
    </div>
  )
}
