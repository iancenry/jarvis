import { FullPageLoader } from "@/components/full-page-loader";
import { ThemeToggle } from "@/components/theme-toggle";
import { Button } from "@/components/ui/button";
import { Sheet, SheetContent, SheetTrigger } from "@/components/ui/sheet";
import { cn } from "@/lib/utils";
import { useAuth, UserButton } from "@clerk/clerk-react";
import {
  createFileRoute,
  Link,
  Navigate,
  Outlet,
  useLocation,
} from "@tanstack/react-router";
import {
  LayoutDashboard,
  CheckSquare,
  Folder,
  Settings,
  Menu,
  Sparkles,
  ChevronRight,
} from "lucide-react";
import { motion } from "motion/react";

export const Route = createFileRoute("/_app")({
  component: AppLayout,
});

const navigation = [
  { name: "Dashboard", href: "/dashboard", icon: LayoutDashboard },
  { name: "Tasks", href: "/todos", icon: CheckSquare },
  { name: "Categories", href: "/categories", icon: Folder },
  { name: "Settings", href: "/settings", icon: Settings },
];

function AppLayout() {
  const { isLoaded, isSignedIn } = useAuth();
  const location = useLocation();

  // Loading state
  if (!isLoaded) {
    return <FullPageLoader message="Loading your workspace..." textured />;
  }

  if (!isSignedIn) {
    return (
      <Navigate to="/sign-in" search={{ redirect: location.href }} replace />
    );
  }

  return (
    <div className="min-h-screen bg-background texture-noise">
      {/* Desktop Sidebar */}
      <aside className="hidden md:fixed md:inset-y-0 md:left-0 md:flex md:w-64 md:flex-col">
        <div className="flex flex-col flex-1 bg-sidebar border-r border-sidebar-border">
          {/* Logo */}
          <div className="flex items-center gap-3 px-6 py-5 border-b border-sidebar-border">
            <div className="w-9 h-9 rounded-xl bg-gradient-to-br from-sidebar-primary to-sidebar-primary/70 flex items-center justify-center shadow-soft">
              <Sparkles className="w-5 h-5 text-sidebar-primary-foreground" />
            </div>
            <span className="text-lg font-semibold text-sidebar-foreground tracking-tight">
              jarvis
            </span>
          </div>

          {/* Navigation */}
          <nav className="flex-1 px-3 py-4 space-y-1 overflow-y-auto">
            {navigation.map((item) => {
              const isActive = location.pathname === item.href;
              return (
                <Link key={item.name} to={item.href}>
                  <motion.div
                    className={cn(
                      "group flex items-center gap-3 px-3 py-2.5 rounded-xl text-sm font-medium transition-all duration-200",
                      isActive
                        ? "bg-sidebar-accent text-sidebar-accent-foreground"
                        : "text-sidebar-foreground/70 hover:text-sidebar-foreground hover:bg-sidebar-accent/50",
                    )}
                    whileHover={{ x: 2 }}
                    whileTap={{ scale: 0.98 }}
                  >
                    <item.icon
                      className={cn(
                        "w-5 h-5 transition-colors",
                        isActive
                          ? "text-sidebar-primary"
                          : "text-sidebar-foreground/50 group-hover:text-sidebar-foreground/70",
                      )}
                    />
                    {item.name}
                    {isActive && (
                      <motion.div
                        layoutId="activeIndicator"
                        className="ml-auto w-1.5 h-1.5 rounded-full bg-sidebar-primary"
                        initial={false}
                        transition={{
                          type: "spring",
                          stiffness: 500,
                          damping: 30,
                        }}
                      />
                    )}
                  </motion.div>
                </Link>
              );
            })}
          </nav>

          {/* Bottom Section */}
          <div className="p-4 border-t border-sidebar-border">
            <div className="flex items-center gap-3">
              <UserButton
                appearance={{
                  elements: {
                    avatarBox: "w-9 h-9 rounded-xl ring-2 ring-sidebar-border",
                  },
                }}
              />
              <div className="flex-1 min-w-0">
                <p className="text-sm font-medium text-sidebar-foreground truncate">
                  Account
                </p>
                <p className="text-xs text-sidebar-foreground/50">
                  Manage settings
                </p>
              </div>
            </div>
          </div>
        </div>
      </aside>

      {/* Mobile Header */}
      <div className="md:hidden">
        <div className="fixed top-0 left-0 right-0 z-50 glass border-b border-border/50 px-4 py-3 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="w-8 h-8 rounded-lg bg-gradient-to-br from-accent to-accent/70 flex items-center justify-center shadow-soft">
              <Sparkles className="w-4 h-4 text-accent-foreground" />
            </div>
            <span className="text-lg font-semibold tracking-tight">jarvis</span>
          </div>
          <Sheet>
            <SheetTrigger asChild>
              <Button variant="ghost" size="icon">
                <Menu className="w-5 h-5" />
              </Button>
            </SheetTrigger>
            <SheetContent
              side="left"
              className="w-72 p-0 bg-sidebar border-sidebar-border"
            >
              <div className="flex flex-col h-full">
                <div className="flex items-center gap-3 px-6 py-5 border-b border-sidebar-border">
                  <div className="w-9 h-9 rounded-xl bg-gradient-to-br from-sidebar-primary to-sidebar-primary/70 flex items-center justify-center">
                    <Sparkles className="w-5 h-5 text-sidebar-primary-foreground" />
                  </div>
                  <span className="text-lg font-semibold text-sidebar-foreground">
                    jarvis
                  </span>
                </div>
                <nav className="flex-1 px-3 py-4 space-y-1">
                  {navigation.map((item) => {
                    const isActive = location.pathname === item.href;
                    return (
                      <Link key={item.name} to={item.href}>
                        <div
                          className={cn(
                            "flex items-center gap-3 px-3 py-2.5 rounded-xl text-sm font-medium transition-colors",
                            isActive
                              ? "bg-sidebar-accent text-sidebar-accent-foreground"
                              : "text-sidebar-foreground/70 hover:text-sidebar-foreground hover:bg-sidebar-accent/50",
                          )}
                        >
                          <item.icon className="w-5 h-5" />
                          {item.name}
                        </div>
                      </Link>
                    );
                  })}
                </nav>
                <div className="p-4 border-t border-sidebar-border">
                  <div className="flex items-center gap-3">
                    <UserButton
                      appearance={{
                        elements: {
                          avatarBox: "w-9 h-9 rounded-xl",
                        },
                      }}
                    />
                    <span className="text-sm text-sidebar-foreground">
                      Account
                    </span>
                  </div>
                </div>
              </div>
            </SheetContent>
          </Sheet>
        </div>
      </div>

      {/* Main Content */}
      <div className="md:pl-64">
        {/* Top Bar */}
        <header className="sticky top-0 z-40 hidden md:flex h-16 items-center gap-4 border-b border-border/50 bg-background/80 backdrop-blur-xl px-6">
          <div className="flex-1">
            <Breadcrumb />
          </div>
          <div className="flex items-center gap-3">
            <ThemeToggle />
            <UserButton
              appearance={{
                elements: {
                  avatarBox: "w-8 h-8 rounded-lg",
                },
              }}
            />
          </div>
        </header>

        {/* Page Content */}
        <main className="min-h-[calc(100vh-4rem)] pt-16 md:pt-0">
          <div className="p-6 md:p-8">
            <Outlet />
          </div>
        </main>
      </div>
    </div>
  );
}

function Breadcrumb() {
  const location = useLocation();
  const currentPage = navigation.find(
    (item) => item.href === location.pathname,
  );

  return (
    <div className="flex items-center gap-2 text-sm">
      <span className="text-muted-foreground">jarvis</span>
      <ChevronRight className="w-4 h-4 text-muted-foreground/50" />
      <span className="font-medium">{currentPage?.name || "Page"}</span>
    </div>
  );
}
