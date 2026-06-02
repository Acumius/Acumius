import {
  Sidebar,
  SidebarContent,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from "@/components/ui/sidebar"
import { Home, Database, Bot, Shield, FileText, Lock } from "lucide-react"

// Menu items.
const items = [
  {
    title: "Dashboard",
    url: "/",
    icon: Home,
  },
  {
    title: "Memory Explorer",
    url: "/memory",
    icon: Database,
  },
  {
    title: "Agent Directory",
    url: "/agents",
    icon: Bot,
  },
  {
    title: "Policies",
    url: "/policies",
    icon: Shield,
  },
  {
    title: "Audit Log",
    url: "/audit",
    icon: FileText,
  },
  {
    title: "GDPR Tools",
    url: "/gdpr",
    icon: Lock,
  },
]

export function AppSidebar() {
  return (
    <Sidebar>
      <SidebarContent>
        <SidebarGroup>
          <SidebarGroupLabel className="flex items-center gap-2 h-14">
            <img src="/logo.png" alt="Acumius Logo" className="h-8 w-8 object-contain" />
            <span className="text-lg font-semibold tracking-tight text-primary">Acumius</span>
          </SidebarGroupLabel>
          <SidebarGroupContent>
            <SidebarMenu>
              {items.map((item) => (
                <SidebarMenuItem key={item.title}>
                  <SidebarMenuButton tooltip={item.title}>
                    <a href={item.url} className="flex items-center gap-2 w-full">
                      <item.icon />
                      <span>{item.title}</span>
                    </a>
                  </SidebarMenuButton>
                </SidebarMenuItem>
              ))}
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>
      </SidebarContent>
    </Sidebar>
  )
}
