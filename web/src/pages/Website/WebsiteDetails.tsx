import { useQuery } from "@tanstack/react-query"
import { useParams, Link } from "react-router-dom"
import { ArrowLeft, Clock, Activity, CheckCircle, XCircle } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import api from "@/lib/api"
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
} from "recharts"

interface WebsiteDetail {
  ID: string
  URL: string
  Ticks: {
    ID: string
    Status: string
    Latency: number
    CreatedAt: string
  }[]
}

export default function WebsiteDetails() {
  const { id } = useParams<{ id: string }>()

  const { data: website, isLoading } = useQuery({
    queryKey: ["website", id],
    queryFn: async () => {
      const res = await api.get<WebsiteDetail>(`/website/status?websiteId=${id}`)
      return res.data
    },
    refetchInterval: 10000, // Poll every 10 seconds
  })

  // Process data for charts
  const chartData = website?.Ticks?.slice().reverse().map(tick => ({
    time: new Date(tick.CreatedAt).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }),
    latency: tick.Latency,
    status: tick.Status
  })) || []

  // Calculate Uptime
  const totalTicks = website?.Ticks?.length || 0
  const goodTicks = website?.Ticks?.filter(t => t.Status === "Good").length || 0
  const uptime = totalTicks > 0 ? ((goodTicks / totalTicks) * 100).toFixed(2) : "0.00"

  const latestStatus = website?.Ticks?.[0]?.Status === "Good"

  if (isLoading) {
    return (
      <div className="container mx-auto p-6 max-w-6xl space-y-6">
        <Skeleton className="h-12 w-48" />
        <div className="grid gap-4 md:grid-cols-3">
          <Skeleton className="h-32" />
          <Skeleton className="h-32" />
          <Skeleton className="h-32" />
        </div>
        <Skeleton className="h-[400px]" />
      </div>
    )
  }

  if (!website) {
    return <div className="p-8 text-center">Website not found</div>
  }

  return (
    <div className="min-h-screen bg-slate-50 dark:bg-slate-900 p-6 md:p-8">
      <div className="container mx-auto max-w-6xl space-y-8">
        <div>
          <Link to="/dashboard">
            <Button variant="ghost" className="mb-4 pl-0 hover:pl-0 hover:bg-transparent hover:text-primary">
              <ArrowLeft className="h-4 w-4 mr-2" />
              Back to Dashboard
            </Button>
          </Link>
          <div className="flex items-center gap-4">
            <h1 className="text-3xl font-bold tracking-tight">{website.URL}</h1>
            <div className={`flex items-center gap-1.5 px-3 py-1 rounded-full text-sm font-medium ${latestStatus ? 'bg-green-100 text-green-700' : 'bg-red-100 text-red-700'}`}>
              {latestStatus ? <CheckCircle className="h-4 w-4" /> : <XCircle className="h-4 w-4" />}
              {latestStatus ? 'Operational' : 'Downtime'}
            </div>
          </div>
        </div>

        <div className="grid gap-4 md:grid-cols-3">
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Uptime (Last 100 ticks)</CardTitle>
              <Activity className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{uptime}%</div>
            </CardContent>
          </Card>
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Latest Latency</CardTitle>
              <Clock className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{website.Ticks?.[0]?.Latency || 0} ms</div>
              <p className="text-xs text-muted-foreground">
                Last checked: {website.Ticks?.[0] ? new Date(website.Ticks[0].CreatedAt).toLocaleTimeString() : 'Never'}
              </p>
            </CardContent>
          </Card>
           <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Monitoring ID</CardTitle>
              <CheckCircle className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-xs font-mono text-muted-foreground break-all">{website.ID}</div>
            </CardContent>
          </Card>
        </div>

        <Card>
          <CardHeader>
            <CardTitle>Latency History</CardTitle>
            <CardDescription>Response time over the last 100 checks</CardDescription>
          </CardHeader>
          <CardContent className="h-[400px]">
            <ResponsiveContainer width="100%" height="100%">
              <LineChart data={chartData}>
                <CartesianGrid strokeDasharray="3 3" vertical={false} />
                <XAxis 
                  dataKey="time" 
                  stroke="#888888" 
                  fontSize={12} 
                  tickLine={false} 
                  axisLine={false} 
                  minTickGap={30}
                />
                <YAxis 
                  stroke="#888888" 
                  fontSize={12} 
                  tickLine={false} 
                  axisLine={false}
                  tickFormatter={(value) => `${value}ms`}
                />
                <Tooltip 
                  contentStyle={{ backgroundColor: 'white', borderRadius: '8px', border: '1px solid #e5e7eb' }}
                  labelStyle={{ color: '#6b7280' }}
                />
                <Line 
                  type="monotone" 
                  dataKey="latency" 
                  stroke="#2563eb" 
                  strokeWidth={2} 
                  dot={false}
                  activeDot={{ r: 4 }} 
                />
              </LineChart>
            </ResponsiveContainer>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
