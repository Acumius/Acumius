import Image from "next/image"
import { Card, CardContent, CardDescription, CardHeader, CardTitle, CardFooter } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"
import { Label } from "@/components/ui/label"
import { AlertTriangle, UserMinus, RefreshCw } from "lucide-react"

export default function GDPRTools() {
  return (
    <div className="flex flex-col gap-8">
      <div className="flex items-center gap-4">
        <Image src="/logo.png" alt="Acumius Logo" width={48} height={48} className="rounded-lg object-contain bg-white p-1" />
        <div>
          <h1 className="text-3xl font-bold tracking-tight">GDPR Compliance Tools</h1>
          <p className="text-muted-foreground mt-2">
            Execute data subjects&apos; rights including Right to Forget and Data Rectification.
          </p>
        </div>
      </div>

      <div className="grid gap-8 md:grid-cols-2">
        <Card className="border-destructive/50">
          <CardHeader>
            <div className="flex items-center space-x-2">
              <UserMinus className="h-5 w-5 text-destructive" />
              <CardTitle>Right to Forget (Erasure)</CardTitle>
            </div>
            <CardDescription>
              Permanently delete all memory records associated with a specific agent DID or User ID.
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="did-forget">Target Agent DID or User ID</Label>
              <Input id="did-forget" placeholder="did:key:..." />
            </div>
            <div className="bg-destructive/10 p-3 rounded-md flex items-start space-x-2">
              <AlertTriangle className="h-5 w-5 text-destructive shrink-0 mt-0.5" />
              <p className="text-sm text-destructive">
                Warning: This action is irreversible. All episodic and working memory for this identifier will be purged from Postgres and Valkey.
              </p>
            </div>
          </CardContent>
          <CardFooter>
            <Button variant="destructive" className="w-full">Execute Erasure Protocol</Button>
          </CardFooter>
        </Card>

        <Card>
          <CardHeader>
            <div className="flex items-center space-x-2">
              <RefreshCw className="h-5 w-5 text-primary" />
              <CardTitle>Data Rectification</CardTitle>
            </div>
            <CardDescription>
              Correct inaccurate or incomplete personal data within a specific memory record.
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="memory-id">Memory ID</Label>
              <Input id="memory-id" placeholder="mem-abc-123" />
            </div>
            <div className="space-y-2">
              <Label htmlFor="corrected-data">Corrected Content (JSON)</Label>
              <textarea 
                id="corrected-data" 
                className="flex min-h-[120px] w-full rounded-md border border-input bg-transparent px-3 py-2 text-sm shadow-sm placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50"
                placeholder='{"content": "updated data"}'
              />
            </div>
          </CardContent>
          <CardFooter>
            <Button className="w-full">Apply Rectification</Button>
          </CardFooter>
        </Card>
      </div>
    </div>
  )
}
