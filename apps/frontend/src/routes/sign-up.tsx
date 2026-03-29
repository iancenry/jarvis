import { FullPageLoader } from "@/components/full-page-loader";
import { ThemeToggle } from "@/components/theme-toggle";
import {
  DEFAULT_AUTH_REDIRECT,
  getAuthRouteUrl,
  getPostAuthRedirect,
  parseAuthSearch,
} from "@/routes/-auth";
import { useAuth } from "@clerk/clerk-react";
import { SignUp } from "@clerk/clerk-react";
import { Navigate, createFileRoute } from "@tanstack/react-router";
import { Sparkles, CheckCircle2 } from "lucide-react";
import { motion } from "motion/react";

export const Route = createFileRoute("/sign-up")({
  validateSearch: parseAuthSearch,
  component: SignUpPage,
});

const features = [
  "Unlimited tasks and categories",
  "Smart organization tools",
  "File attachments support",
  "Real-time collaboration",
];

function SignUpPage() {
  const { isLoaded, isSignedIn } = useAuth();
  const { redirect: redirectTo } = Route.useSearch();

  if (!isLoaded) {
    return <FullPageLoader message="Checking your session..." />;
  }

  if (isSignedIn) {
    return <Navigate to="/" href={getPostAuthRedirect(redirectTo)} replace />;
  }

  const redirectProps = redirectTo
    ? { forceRedirectUrl: redirectTo }
    : { fallbackRedirectUrl: DEFAULT_AUTH_REDIRECT };

  return (
    <div className="min-h-screen flex texture-noise">
      {/* Left side - Branding */}
      <div className="hidden lg:flex lg:w-1/2 bg-sidebar flex-col justify-between p-12">
        <motion.div
          className="flex items-center gap-3"
          initial={{ opacity: 0, y: -20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.5 }}
        >
          <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-sidebar-primary to-sidebar-primary/70 flex items-center justify-center shadow-soft">
            <Sparkles className="w-5 h-5 text-sidebar-primary-foreground" />
          </div>
          <span className="text-xl font-semibold text-sidebar-foreground tracking-tight">
            jarvis
          </span>
        </motion.div>

        <motion.div
          className="space-y-8"
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.6, delay: 0.2 }}
        >
          <div className="space-y-4">
            <h1 className="text-4xl xl:text-5xl font-bold text-sidebar-foreground leading-tight">
              Start your journey<span className="text-sidebar-primary">.</span>
            </h1>
            <p className="text-lg text-sidebar-foreground/70 max-w-md leading-relaxed">
              Create your free account and start organizing your tasks like
              never before.
            </p>
          </div>

          <div className="space-y-3 stagger-in">
            {features.map((feature, index) => (
              <motion.div
                key={feature}
                className="flex items-center gap-3"
                initial={{ opacity: 0, x: -20 }}
                animate={{ opacity: 1, x: 0 }}
                transition={{ duration: 0.4, delay: 0.3 + index * 0.1 }}
              >
                <div className="w-6 h-6 rounded-full bg-sidebar-primary/20 flex items-center justify-center">
                  <CheckCircle2 className="w-4 h-4 text-sidebar-primary" />
                </div>
                <span className="text-sidebar-foreground/80">{feature}</span>
              </motion.div>
            ))}
          </div>
        </motion.div>

        <motion.div
          className="flex items-center gap-4 text-sm text-sidebar-foreground/50"
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          transition={{ duration: 0.5, delay: 0.6 }}
        >
          <span>&copy; {new Date().getFullYear()} Jarvis</span>
          <span className="w-1 h-1 rounded-full bg-sidebar-foreground/30" />
          <span>Privacy Policy</span>
          <span className="w-1 h-1 rounded-full bg-sidebar-foreground/30" />
          <span>Terms of Service</span>
        </motion.div>
      </div>

      {/* Right side - Sign Up Form */}
      <div className="flex-1 flex flex-col">
        {/* Mobile header */}
        <div className="lg:hidden flex items-center justify-between p-4 border-b border-border/50">
          <div className="flex items-center gap-2">
            <div className="w-8 h-8 rounded-lg bg-gradient-to-br from-accent to-accent/70 flex items-center justify-center">
              <Sparkles className="w-4 h-4 text-accent-foreground" />
            </div>
            <span className="font-semibold">jarvis</span>
          </div>
          <ThemeToggle />
        </div>

        <div className="flex-1 flex items-center justify-center p-6">
          <motion.div
            initial={{ opacity: 0, scale: 0.95 }}
            animate={{ opacity: 1, scale: 1 }}
            transition={{ duration: 0.4 }}
            className="w-full max-w-md"
          >
            <div className="lg:hidden text-center mb-8">
              <h1 className="text-2xl font-bold">Create your account</h1>
              <p className="text-muted-foreground mt-1">
                Start organizing your tasks today
              </p>
            </div>
            <SignUp
              appearance={{
                elements: {
                  rootBox: "w-full",
                  card: "shadow-soft rounded-2xl border border-border/50 bg-card",
                  headerTitle: "text-xl font-semibold",
                  headerSubtitle: "text-muted-foreground",
                  socialButtonsBlockButton:
                    "border-border hover:bg-muted/50 transition-colors",
                  formFieldInput:
                    "rounded-xl border-input bg-muted/30 focus:border-accent focus:ring-1 focus:ring-accent/30",
                  formButtonPrimary:
                    "bg-primary hover:bg-primary/90 rounded-xl shadow-soft",
                  footerActionLink: "text-accent hover:text-accent/80",
                },
              }}
              routing="path"
              path="/sign-up"
              signInUrl={getAuthRouteUrl("/sign-in", redirectTo)}
              {...redirectProps}
            />
          </motion.div>
        </div>

        {/* Desktop theme toggle */}
        <div className="hidden lg:flex justify-end p-6">
          <ThemeToggle />
        </div>
      </div>
    </div>
  );
}
