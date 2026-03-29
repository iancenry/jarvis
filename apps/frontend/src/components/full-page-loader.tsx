import { cn } from "@/lib/utils";
import { Sparkles } from "lucide-react";
import { motion } from "motion/react";

interface FullPageLoaderProps {
  message?: string;
  textured?: boolean;
  className?: string;
}

export function FullPageLoader({
  message = "Loading...",
  textured = false,
  className,
}: FullPageLoaderProps) {
  return (
    <div
      className={cn(
        "min-h-screen flex items-center justify-center bg-background",
        textured && "texture-noise",
        className,
      )}
    >
      <motion.div
        className="flex flex-col items-center gap-4"
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
      >
        <div className="w-12 h-12 rounded-xl bg-gradient-to-br from-accent to-accent/70 flex items-center justify-center shadow-soft animate-pulse">
          <Sparkles className="w-6 h-6 text-accent-foreground" />
        </div>
        <p className="text-muted-foreground text-sm">{message}</p>
      </motion.div>
    </div>
  );
}
