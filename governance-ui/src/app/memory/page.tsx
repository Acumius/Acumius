import Image from "next/image"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Search, Filter, Database } from "lucide-react"

export default function MemoryExplorer() {
  return (
    <div className="flex flex-col gap-8">
      <div className="flex items-center gap-4">
        <Image src="/logo.png" alt="Acumius Logo" width={48} height={48} className="rounded-lg object-contain bg-white p-1" />
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Memory Explorer</h1>
          <p className="text-muted-foreground mt-2">
            Search and filter across all namespace memories.
          </p>
        </div>
      </div>

      <div className="flex w-full items-center space-x-2">
        <Input type="text" placeholder="Search memories (Hybrid vector search)..." className="max-w-xl" />
        <Button type="submit">
          <Search className="mr-2 h-4 w-4" />
          Search
        </Button>
        <Button variant="outline">
          <Filter className="mr-2 h-4 w-4" />
          Filter
        </Button>
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        {[1, 2, 3, 4, 5, 6].map((i) => (
          <Card key={i}>
            <CardHeader className="pb-3">
              <div className="flex items-center justify-between">
                <Badge variant={i % 2 === 0 ? "default" : "secondary"}>
                  {i % 2 === 0 ? "Episodic" : "Working"}
                </Badge>
                <Database className="h-4 w-4 text-muted-foreground" />
              </div>
              <CardTitle className="text-lg mt-4">User interaction memory {i}</CardTitle>
              <CardDescription>Agent: did:key:z6Mk...</CardDescription>
            </CardHeader>
            <CardContent>
              <p className="text-sm text-muted-foreground line-clamp-3">
                This is a placeholder for the actual memory content. It stores the relevant interactions and contextual information that the agent needs to recall later for personalized assistance and task execution.
              </p>
              <div className="mt-4 text-xs text-muted-foreground">
                Score: 0.9{i} • 2 hours ago
              </div>
            </CardContent>
          </Card>
        ))}
      </div>
    </div>
  )
}
