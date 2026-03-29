import { useAuth } from "@clerk/clerk-react";
import { useNavigate } from "@tanstack/react-router";
import { Sparkles } from "lucide-react";
import { type ReactNode } from "react";
import { useEffect } from "react";

interface PublicRouteProps {
  children: ReactNode;
}

export function PublicRoute({ children }: PublicRouteProps) {
  const { isLoaded, isSignedIn } = useAuth();
  const navigate = useNavigate();

  useEffect(() => {
    if (isLoaded && isSignedIn) {
      navigate({ to: "/dashboard" });
    }
  }, [isLoaded, isSignedIn, navigate]);

  if (!isLoaded) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-background">
        <div className="flex flex-col items-center gap-4">
          <div className="w-12 h-12 rounded-xl bg-gradient-to-br from-accent to-accent/70 flex items-center justify-center shadow-soft animate-pulse">
            <Sparkles className="w-6 h-6 text-accent-foreground" />
          </div>
          <p className="text-muted-foreground text-sm">Loading...</p>
        </div>
      </div>
    );
  }

  if (isSignedIn) {
    return null;
  }

  return <>{children}</>;
}
