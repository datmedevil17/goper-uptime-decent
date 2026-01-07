import { Button } from "@/components/ui/button"
import { ExternalLink, ShieldCheck, Activity, Globe } from "lucide-react"
import { Link } from "react-router-dom"

export default function LandingPage() {
  return (
    <div className="flex flex-col min-h-screen">
      <header className="px-4 lg:px-6 h-14 flex items-center border-b">
        <Link className="flex items-center justify-center" to="/">
          <Activity className="h-6 w-6 mr-2" />
          <span className="font-bold text-xl">GopherUptime</span>
        </Link>
        <nav className="ml-auto flex gap-4 sm:gap-6">
          <Link className="text-sm font-medium hover:underline underline-offset-4" to="/login">
            Login
          </Link>
          <Link className="text-sm font-medium hover:underline underline-offset-4" to="/signup">
            Sign Up
          </Link>
        </nav>
      </header>
      <main className="flex-1">
        <section className="w-full py-12 md:py-24 lg:py-32 xl:py-48 bg-slate-50 dark:bg-slate-900">
          <div className="container px-4 md:px-6">
            <div className="flex flex-col items-center space-y-4 text-center">
              <div className="space-y-2">
                <h1 className="text-3xl font-bold tracking-tighter sm:text-4xl md:text-5xl lg:text-6xl/none">
                  Decentralized Uptime Monitoring
                </h1>
                <p className="mx-auto max-w-[700px] text-gray-500 md:text-xl dark:text-gray-400">
                  Monitor your websites with a distributed network of validators. Earn rewards for helping secure the web.
                </p>
              </div>
              <div className="space-x-4">
                <Link to="/signup">
                  <Button size="lg">Get Started</Button>
                </Link>
                <Link to="/login">
                  <Button variant="outline" size="lg">Login</Button>
                </Link>
              </div>
            </div>
          </div>
        </section>
        <section className="w-full py-12 md:py-24 lg:py-32">
          <div className="container px-4 md:px-6">
            <div className="grid gap-10 sm:grid-cols-2 md:order-2 lg:grid-cols-3">
              <div className="flex flex-col items-center space-y-2 border-gray-800 p-4 rounded-lg">
                <div className="p-2 bg-slate-100 rounded-full dark:bg-slate-800">
                  <Globe className="h-6 w-6" />
                </div>
                <h2 className="text-xl font-bold">Global Coverage</h2>
                <p className="text-gray-500 dark:text-gray-400 text-center">
                  Validators distributed worldwide ensuring accurate latency checks from multiple regions.
                </p>
              </div>
              <div className="flex flex-col items-center space-y-2 border-gray-800 p-4 rounded-lg">
                <div className="p-2 bg-slate-100 rounded-full dark:bg-slate-800">
                  <ShieldCheck className="h-6 w-6" />
                </div>
                <h2 className="text-xl font-bold">Cryptographically Verified</h2>
                <p className="text-gray-500 dark:text-gray-400 text-center">
                  Every check is signed by a validator's private key, ensuring data integrity.
                </p>
              </div>
              <div className="flex flex-col items-center space-y-2 border-gray-800 p-4 rounded-lg">
                <div className="p-2 bg-slate-100 rounded-full dark:bg-slate-800">
                  <ExternalLink className="h-6 w-6" />
                </div>
                <h2 className="text-xl font-bold">Earn Rewards</h2>
                <p className="text-gray-500 dark:text-gray-400 text-center">
                  Run a validator node and get paid for performing health checks.
                </p>
              </div>
            </div>
          </div>
        </section>
      </main>
      <footer className="flex flex-col gap-2 sm:flex-row py-6 w-full shrink-0 items-center px-4 md:px-6 border-t">
        <p className="text-xs text-gray-500 dark:text-gray-400">© 2024 GopherUptime. All rights reserved.</p>
        <nav className="sm:ml-auto flex gap-4 sm:gap-6">
          <Link className="text-xs hover:underline underline-offset-4" to="#">
            Terms of Service
          </Link>
          <Link className="text-xs hover:underline underline-offset-4" to="#">
            Privacy
          </Link>
        </nav>
      </footer>
    </div>
  )
}
