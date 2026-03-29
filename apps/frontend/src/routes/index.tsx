import { ThemeToggle } from "@/components/theme-toggle";
import { Button } from "@/components/ui/button";
import { useAuth } from "@clerk/clerk-react";
import { createFileRoute, Link, useNavigate } from "@tanstack/react-router";
import {
  CheckCircle2,
  Users2,
  Zap,
  Shield,
  ArrowRight,
  Sparkles,
} from "lucide-react";
import { motion } from "motion/react";
import { useEffect } from "react";

export const Route = createFileRoute("/")({
  component: LandingPage,
});

function LandingPage() {
  const { isSignedIn, isLoaded } = useAuth();
  const navigate = useNavigate();

  // Redirect to dashboard if already signed in
  useEffect(() => {
    if (isLoaded && isSignedIn) {
      navigate({ to: "/dashboard" });
    }
  }, [isLoaded, isSignedIn, navigate]);

  if (isLoaded && isSignedIn) {
    return null;
  }

  return (
    <div className="min-h-screen bg-background texture-noise">
      {/* Header */}
      <header className="fixed top-0 left-0 right-0 z-50 glass border-b border-border/50">
        <div className="container mx-auto px-6 py-4 flex items-center justify-between">
          <motion.div
            className="flex items-center gap-3"
            initial={{ opacity: 0, x: -20 }}
            animate={{ opacity: 1, x: 0 }}
            transition={{ duration: 0.5 }}
          >
            <div className="w-9 h-9 rounded-xl bg-gradient-to-br from-accent to-accent/70 flex items-center justify-center shadow-soft">
              <Sparkles className="w-5 h-5 text-accent-foreground" />
            </div>
            <span className="text-xl font-semibold tracking-tight">jarvis</span>
          </motion.div>

          <motion.div
            className="flex items-center gap-3"
            initial={{ opacity: 0, x: 20 }}
            animate={{ opacity: 1, x: 0 }}
            transition={{ duration: 0.5 }}
          >
            <ThemeToggle />
            <Link to="/sign-in">
              <Button variant="ghost" className="font-medium">
                Sign In
              </Button>
            </Link>
            <Link to="/sign-up">
              <Button className="font-medium shadow-soft press-scale">
                Get Started
              </Button>
            </Link>
          </motion.div>
        </div>
      </header>

      {/* Hero Section */}
      <section className="pt-32 pb-20 px-6 overflow-hidden">
        <div className="container mx-auto max-w-5xl">
          <div className="text-center space-y-8">
            <motion.div
              initial={{ opacity: 0, y: 30 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.7, delay: 0.1 }}
            >
              <span className="inline-flex items-center gap-2 px-4 py-2 rounded-full bg-accent/10 text-accent text-sm font-medium border border-accent/20">
                <Sparkles className="w-4 h-4" />
                Intelligent task management
              </span>
            </motion.div>

            <motion.h1
              className="text-5xl md:text-7xl font-bold tracking-tight leading-[1.1]"
              initial={{ opacity: 0, y: 30 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.7, delay: 0.2 }}
            >
              <span className="text-gradient">Organize</span> your work.
              <br />
              <span className="text-muted-foreground">Amplify</span> your focus.
            </motion.h1>

            <motion.p
              className="text-xl text-muted-foreground max-w-2xl mx-auto leading-relaxed"
              initial={{ opacity: 0, y: 30 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.7, delay: 0.3 }}
            >
              A beautifully crafted task manager that adapts to your workflow.
              Stay productive with smart organization, seamless collaboration,
              and powerful insights.
            </motion.p>

            <motion.div
              className="flex flex-col sm:flex-row gap-4 justify-center pt-4"
              initial={{ opacity: 0, y: 30 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.7, delay: 0.4 }}
            >
              <Link to="/sign-up">
                <Button
                  size="lg"
                  className="text-base px-8 py-6 shadow-elevated hover-lift press-scale group"
                >
                  Start for free
                  <ArrowRight className="ml-2 w-4 h-4 transition-transform group-hover:translate-x-1" />
                </Button>
              </Link>
              <Link to="/sign-in">
                <Button
                  variant="outline"
                  size="lg"
                  className="text-base px-8 py-6 hover-lift press-scale"
                >
                  Sign in to your account
                </Button>
              </Link>
            </motion.div>
          </div>

          {/* Hero Visual */}
          <motion.div
            className="mt-20 relative"
            initial={{ opacity: 0, y: 50 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.8, delay: 0.5 }}
          >
            <div className="absolute inset-0 bg-gradient-to-t from-background via-transparent to-transparent z-10" />
            <div className="rounded-2xl border border-border/50 shadow-elevated overflow-hidden bg-card">
              <div className="p-1 bg-gradient-to-b from-muted/50 to-transparent">
                <div className="flex items-center gap-1.5 px-3 py-2">
                  <div className="w-2.5 h-2.5 rounded-full bg-destructive/60" />
                  <div className="w-2.5 h-2.5 rounded-full bg-accent/60" />
                  <div className="w-2.5 h-2.5 rounded-full bg-chart-2/60" />
                </div>
              </div>
              <div className="aspect-[16/9] bg-gradient-to-br from-muted/30 via-card to-muted/50 flex items-center justify-center">
                <div className="text-center space-y-4">
                  <div className="w-16 h-16 mx-auto rounded-2xl bg-gradient-to-br from-accent to-accent/70 flex items-center justify-center shadow-soft">
                    <CheckCircle2 className="w-8 h-8 text-accent-foreground" />
                  </div>
                  <p className="text-muted-foreground font-medium">
                    Your productivity dashboard awaits
                  </p>
                </div>
              </div>
            </div>
          </motion.div>
        </div>
      </section>

      {/* Features Section */}
      <section className="py-24 px-6 bg-muted/30">
        <div className="container mx-auto max-w-6xl">
          <motion.div
            className="text-center mb-16"
            initial={{ opacity: 0, y: 20 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={{ once: true }}
            transition={{ duration: 0.6 }}
          >
            <h2 className="text-3xl md:text-4xl font-bold mb-4">
              Everything you need to stay organized
            </h2>
            <p className="text-lg text-muted-foreground max-w-2xl mx-auto">
              Powerful features wrapped in a simple, intuitive interface
            </p>
          </motion.div>

          <div className="grid md:grid-cols-2 lg:grid-cols-4 gap-6 stagger-in">
            <FeatureCard
              icon={CheckCircle2}
              title="Smart Organization"
              description="Organize tasks with categories, priorities, and custom due dates"
              delay={0}
            />
            <FeatureCard
              icon={Users2}
              title="Collaboration"
              description="Add comments and work together on shared projects"
              delay={0.1}
            />
            <FeatureCard
              icon={Zap}
              title="Lightning Fast"
              description="Quick actions and keyboard shortcuts for power users"
              delay={0.2}
            />
            <FeatureCard
              icon={Shield}
              title="Secure & Private"
              description="Your data is encrypted and securely stored"
              delay={0.3}
            />
          </div>
        </div>
      </section>

      {/* CTA Section */}
      <section className="py-24 px-6">
        <div className="container mx-auto max-w-3xl">
          <motion.div
            className="text-center space-y-8 p-12 rounded-3xl bg-gradient-to-br from-card to-muted/30 border border-border/50 shadow-elevated"
            initial={{ opacity: 0, scale: 0.95 }}
            whileInView={{ opacity: 1, scale: 1 }}
            viewport={{ once: true }}
            transition={{ duration: 0.6 }}
          >
            <div className="w-14 h-14 mx-auto rounded-2xl bg-gradient-to-br from-accent to-accent/70 flex items-center justify-center shadow-soft glow-accent">
              <Sparkles className="w-7 h-7 text-accent-foreground" />
            </div>
            <h2 className="text-3xl md:text-4xl font-bold">
              Ready to get organized?
            </h2>
            <p className="text-lg text-muted-foreground max-w-xl mx-auto">
              Join thousands of users who have transformed their productivity
              with jarvis.
            </p>
            <Link to="/sign-up">
              <Button
                size="lg"
                className="text-base px-8 py-6 shadow-soft hover-lift press-scale group"
              >
                Start your journey
                <ArrowRight className="ml-2 w-4 h-4 transition-transform group-hover:translate-x-1" />
              </Button>
            </Link>
          </motion.div>
        </div>
      </section>

      {/* Footer */}
      <footer className="py-8 px-6 border-t border-border/50">
        <div className="container mx-auto flex flex-col sm:flex-row items-center justify-between gap-4">
          <div className="flex items-center gap-2">
            <div className="w-7 h-7 rounded-lg bg-gradient-to-br from-accent to-accent/70 flex items-center justify-center">
              <Sparkles className="w-4 h-4 text-accent-foreground" />
            </div>
            <span className="font-semibold">jarvis</span>
          </div>
          <p className="text-sm text-muted-foreground">
            &copy; {new Date().getFullYear()} Jarvis. All rights reserved.
          </p>
        </div>
      </footer>
    </div>
  );
}

interface FeatureCardProps {
  icon: React.ElementType;
  title: string;
  description: string;
  delay: number;
}

function FeatureCard({
  icon: Icon,
  title,
  description,
  delay,
}: FeatureCardProps) {
  return (
    <motion.div
      className="group p-6 rounded-2xl bg-card border border-border/50 shadow-soft hover-lift transition-all duration-300 hover:border-accent/30"
      initial={{ opacity: 0, y: 20 }}
      whileInView={{ opacity: 1, y: 0 }}
      viewport={{ once: true }}
      transition={{ duration: 0.5, delay }}
    >
      <div className="w-12 h-12 rounded-xl bg-gradient-to-br from-accent/10 to-accent/5 flex items-center justify-center mb-4 group-hover:from-accent/20 group-hover:to-accent/10 transition-colors">
        <Icon className="w-6 h-6 text-accent" />
      </div>
      <h3 className="text-lg font-semibold mb-2">{title}</h3>
      <p className="text-muted-foreground text-sm leading-relaxed">
        {description}
      </p>
    </motion.div>
  );
}
