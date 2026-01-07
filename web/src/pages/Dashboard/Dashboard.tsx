import { useState } from "react"
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query"
import { useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import * as z from "zod"
import { Link } from "react-router-dom"
import { Plus, ExternalLink, Activity, Trash2 } from "lucide-react"
import { Button } from "@/components/ui/button"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog"
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form"
import { Input } from "@/components/ui/input"

// Shadcn usually has badge. I didn't install it. I'll use standard tailwind classes.
import api from "@/lib/api"
import { toast } from "sonner"
import { useAuth } from "@/context/AuthContext"

// Types
interface Website {
  ID: string
  URL: string
  Ticks: {
    ID: string
    Status: string
    Latency: number
    CreatedAt: string
  }[]
}

interface WebsitesResponse {
  websites: Website[]
  count: number
}

const formSchema = z.object({
  url: z.string().url("Please enter a valid URL"),
})

export default function Dashboard() {
  const { user, logout } = useAuth()
  const [open, setOpen] = useState(false)
  const queryClient = useQueryClient()

  // Fetch Websites
  const { data, isLoading } = useQuery({
    queryKey: ["websites"],
    queryFn: async () => {
      const res = await api.get<WebsitesResponse>("/websites")
      return res.data
    },
  })

  // Add Website Mutation
  const mutation = useMutation({
    mutationFn: (values: z.infer<typeof formSchema>) => {
      return api.post("/website", values)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["websites"] })
      toast.success("Website added successfully")
      setOpen(false)
      form.reset()
    },
    onError: (error: any) => {
      toast.error("Failed to add website", {
        description: error.response?.data?.error || "Unknown error",
      })
    },
  })

  // Delete Website Mutation
  const deleteMutation = useMutation({
    mutationFn: (websiteId: string) => {
      return api.delete("/website", { data: { websiteId } })
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["websites"] })
      toast.success("Website deleted")
    },
    onError: (_error: any) => {
      toast.error("Failed to delete website")
    },
  })

  const form = useForm<z.infer<typeof formSchema>>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      url: "",
    },
  })

  function onSubmit(values: z.infer<typeof formSchema>) {
    mutation.mutate(values)
  }

  return (
    <div className="flex flex-col min-h-screen bg-slate-50 dark:bg-slate-900">
      <header className="px-6 h-16 flex items-center border-b bg-white dark:bg-slate-950 sticky top-0 z-10">
        <div className="flex items-center gap-2 font-bold text-xl">
          <Activity className="h-6 w-6 text-primary" />
          GopherUptime
        </div>
        <div className="ml-auto flex items-center gap-4">
          <span className="text-sm text-gray-500">{user?.email}</span>
          <Button variant="outline" size="sm" onClick={logout}>Logout</Button>
        </div>
      </header>
      
      <main className="flex-1 p-6 md:p-8 container mx-auto max-w-6xl">
        <div className="flex items-center justify-between mb-8">
          <div>
            <h1 className="text-3xl font-bold tracking-tight">Dashboard</h1>
            <p className="text-gray-500">Manage your monitored websites.</p>
          </div>
          <Dialog open={open} onOpenChange={setOpen}>
            <DialogTrigger asChild>
              <Button>
                <Plus className="h-4 w-4 mr-2" />
                Add Website
              </Button>
            </DialogTrigger>
            <DialogContent>
              <DialogHeader>
                <DialogTitle>Add Website</DialogTitle>
                <DialogDescription>
                  Enter the URL of the website you want to monitor.
                </DialogDescription>
              </DialogHeader>
              <Form {...form}>
                <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
                  <FormField
                    control={form.control}
                    name="url"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>URL</FormLabel>
                        <FormControl>
                          <Input placeholder="https://example.com" {...field} />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                  <Button type="submit" className="w-full" disabled={mutation.isPending}>
                    {mutation.isPending ? "Adding..." : "Add Website"}
                  </Button>
                </form>
              </Form>
            </DialogContent>
          </Dialog>
        </div>

        <div className="bg-white dark:bg-slate-950 rounded-lg border shadow-sm">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Status</TableHead>
                <TableHead>URL</TableHead>
                <TableHead>Latency</TableHead>
                <TableHead>Last Checked</TableHead>
                <TableHead className="text-right">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {isLoading ? (
                <TableRow>
                   <TableCell colSpan={5} className="h-24 text-center">Loading...</TableCell>
                </TableRow>
              ) : data?.websites?.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={5} className="h-24 text-center">No websites found. Add one to get started.</TableCell>
                </TableRow>
              ) : (
                data?.websites?.map((site) => {
                  const lastTick = site.Ticks?.[0]
                  const isUp = lastTick?.Status === "Good"
                  
                  return (
                    <TableRow key={site.ID}>
                      <TableCell>
                        <div className={`h-2.5 w-2.5 rounded-full ${isUp ? 'bg-green-500' : 'bg-red-500'}`} title={lastTick?.Status || "Unknown"} />
                      </TableCell>
                      <TableCell className="font-medium">
                        <Link to={`/website/${site.ID}`} className="hover:underline flex items-center gap-1">
                          {site.URL}
                        </Link>
                      </TableCell>
                      <TableCell>
                        {lastTick ? `${lastTick.Latency}ms` : "-"}
                      </TableCell>
                      <TableCell>
                        {lastTick ? new Date(lastTick.CreatedAt).toLocaleString() : "Never"}
                      </TableCell>
                      <TableCell className="text-right">
                        <Button 
                          variant="ghost" 
                          size="icon" 
                          onClick={() => window.open(site.URL, '_blank')}
                          title="Visit Website"
                        >
                          <ExternalLink className="h-4 w-4" />
                        </Button>
                        <Button
                          variant="ghost" 
                          size="icon"
                          className="text-red-500 hover:text-red-700 hover:bg-red-50"
                          onClick={() => {
                            if (confirm("Are you sure you want to delete this website?")) {
                              deleteMutation.mutate(site.ID)
                            }
                          }}
                        >
                          <Trash2 className="h-4 w-4" />
                        </Button>
                      </TableCell>
                    </TableRow>
                  )
                })
              )}
            </TableBody>
          </Table>
        </div>
      </main>
    </div>
  )
}
