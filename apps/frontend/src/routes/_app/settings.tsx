import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import { cn } from "@/lib/utils";
import { useUser } from "@clerk/clerk-react";
import { createFileRoute } from "@tanstack/react-router";
import {
  User,
  Palette,
  Shield,
  ExternalLink,
  Sun,
  Moon,
  Monitor,
  Check,
} from "lucide-react";
import { motion } from "motion/react";
import { useTheme } from "next-themes";

export const Route = createFileRoute("/_app/settings")({
  component: SettingsPage,
});

function SettingsPage() {
  const { user } = useUser();
  const { theme, setTheme } = useTheme();
  const activeTheme = theme ?? "system";

  const themes = [
    { value: "light", label: "Light", icon: Sun },
    { value: "dark", label: "Dark", icon: Moon },
    { value: "system", label: "System", icon: Monitor },
  ] as const;

  return (
    <div className="space-y-8 max-w-4xl">
      {/* Header */}
      <motion.div
        initial={{ opacity: 0, y: 10 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.4 }}
      >
        <h1 className="text-3xl md:text-4xl font-bold tracking-tight">
          Settings
        </h1>
        <p className="text-muted-foreground mt-1">
          Manage your account and application preferences
        </p>
      </motion.div>

      <div className="space-y-6">
        {/* Profile Settings */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.4, delay: 0.1 }}
        >
          <Card className="shadow-soft">
            <CardHeader className="border-b border-border/50">
              <CardTitle className="flex items-center gap-3 text-lg">
                <div className="w-9 h-9 rounded-xl bg-gradient-to-br from-accent/20 to-accent/5 flex items-center justify-center">
                  <User className="w-5 h-5 text-accent" />
                </div>
                Profile
              </CardTitle>
            </CardHeader>
            <CardContent className="p-6 space-y-6">
              <div className="flex items-start gap-5">
                <div className="relative">
                  <img
                    src={user?.imageUrl}
                    alt={user?.fullName || "User"}
                    className="w-20 h-20 rounded-2xl shadow-soft object-cover"
                  />
                  <div className="absolute -bottom-1 -right-1 w-5 h-5 rounded-full bg-accent flex items-center justify-center ring-2 ring-background">
                    <Check className="w-3 h-3 text-accent-foreground" />
                  </div>
                </div>
                <div className="flex-1">
                  <h3 className="text-xl font-semibold">{user?.fullName}</h3>
                  <p className="text-muted-foreground mt-0.5">
                    {user?.primaryEmailAddress?.emailAddress}
                  </p>
                  <Badge
                    variant="secondary"
                    className="mt-2 font-medium bg-accent/10 text-accent border-accent/20"
                  >
                    {(user?.publicMetadata?.role as string) || "User"}
                  </Badge>
                </div>
              </div>

              <Separator />

              <div className="grid sm:grid-cols-2 gap-4">
                <div className="p-4 rounded-xl bg-muted/30">
                  <p className="text-sm text-muted-foreground">Member since</p>
                  <p className="font-medium mt-1">
                    {user?.createdAt
                      ? new Date(user.createdAt).toLocaleDateString("en-US", {
                          year: "numeric",
                          month: "long",
                          day: "numeric",
                        })
                      : "N/A"}
                  </p>
                </div>
                <div className="p-4 rounded-xl bg-muted/30">
                  <p className="text-sm text-muted-foreground">Last updated</p>
                  <p className="font-medium mt-1">
                    {user?.updatedAt
                      ? new Date(user.updatedAt).toLocaleDateString("en-US", {
                          year: "numeric",
                          month: "long",
                          day: "numeric",
                        })
                      : "N/A"}
                  </p>
                </div>
              </div>

              <Button
                variant="outline"
                className="w-full sm:w-auto press-scale"
                onClick={() =>
                  window.open("https://accounts.clerk.dev", "_blank")
                }
              >
                <ExternalLink className="w-4 h-4 mr-2" />
                Manage Account in Clerk
              </Button>
            </CardContent>
          </Card>
        </motion.div>

        {/* Appearance Settings */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.4, delay: 0.15 }}
        >
          <Card className="shadow-soft">
            <CardHeader className="border-b border-border/50">
              <CardTitle className="flex items-center gap-3 text-lg">
                <div className="w-9 h-9 rounded-xl bg-gradient-to-br from-chart-2/20 to-chart-2/5 flex items-center justify-center">
                  <Palette className="w-5 h-5 text-chart-2" />
                </div>
                Appearance
              </CardTitle>
            </CardHeader>
            <CardContent className="p-6 space-y-6">
              <div>
                <p className="font-medium mb-1">Theme</p>
                <p className="text-sm text-muted-foreground mb-4">
                  Choose how the application looks to you
                </p>
                <div className="grid grid-cols-3 gap-3">
                  {themes.map((t) => {
                    const Icon = t.icon;
                    const isActive = activeTheme === t.value;
                    return (
                      <button
                        key={t.value}
                        type="button"
                        onClick={() => setTheme(t.value)}
                        className={cn(
                          "relative p-4 rounded-xl border-2 transition-all duration-200 press-scale",
                          isActive
                            ? "border-accent bg-accent/5"
                            : "border-border hover:border-accent/50",
                        )}
                      >
                        {isActive && (
                          <div className="absolute top-2 right-2 w-4 h-4 rounded-full bg-accent flex items-center justify-center">
                            <Check className="w-2.5 h-2.5 text-accent-foreground" />
                          </div>
                        )}
                        <Icon
                          className={cn(
                            "w-6 h-6 mx-auto mb-2",
                            isActive ? "text-accent" : "text-muted-foreground",
                          )}
                        />
                        <p
                          className={cn(
                            "text-sm font-medium",
                            isActive ? "text-accent" : "text-foreground",
                          )}
                        >
                          {t.label}
                        </p>
                      </button>
                    );
                  })}
                </div>
              </div>
            </CardContent>
          </Card>
        </motion.div>

        {/* Data & Privacy */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.4, delay: 0.2 }}
        >
          <Card className="shadow-soft">
            <CardHeader className="border-b border-border/50">
              <CardTitle className="flex items-center gap-3 text-lg">
                <div className="w-9 h-9 rounded-xl bg-gradient-to-br from-chart-5/20 to-chart-5/5 flex items-center justify-center">
                  <Shield className="w-5 h-5 text-chart-5" />
                </div>
                Data & Privacy
              </CardTitle>
            </CardHeader>
            <CardContent className="p-6 space-y-4">
              <div className="flex items-center justify-between p-4 rounded-xl bg-muted/30">
                <div>
                  <p className="font-medium">Data Storage</p>
                  <p className="text-sm text-muted-foreground">
                    Your data is securely stored and encrypted
                  </p>
                </div>
                <Badge
                  variant="secondary"
                  className="bg-chart-2/10 text-chart-2 border-chart-2/20"
                >
                  Secure
                </Badge>
              </div>
              <div className="flex items-center justify-between p-4 rounded-xl bg-muted/30">
                <div>
                  <p className="font-medium">Authentication</p>
                  <p className="text-sm text-muted-foreground">
                    Powered by Clerk for secure sign-in
                  </p>
                </div>
                <Badge
                  variant="secondary"
                  className="bg-chart-2/10 text-chart-2 border-chart-2/20"
                >
                  Protected
                </Badge>
              </div>
            </CardContent>
          </Card>
        </motion.div>
      </div>
    </div>
  );
}
