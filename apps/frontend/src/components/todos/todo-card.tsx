import { TodoCommentsDialog } from "./todo-comments-dialog";
import { TodoEditForm } from "./todo-edit-form";
import {
  useUpdateTodo,
  useDeleteTodo,
  type TGetTodosResponse,
} from "@/api/hooks/use-todo-query";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Checkbox } from "@/components/ui/checkbox";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { cn } from "@/lib/utils";
import {
  MoreHorizontal,
  Calendar,
  MessageSquare,
  Edit,
  Trash2,
  Clock,
  CheckCircle2,
  Circle,
  AlertTriangle,
  Archive,
  Paperclip,
  Flame,
} from "lucide-react";
import { motion } from "motion/react";
import { useState } from "react";
import { toast } from "sonner";

interface TodoCardProps {
  todo: TGetTodosResponse["data"][number];
  compact?: boolean;
}

const priorityConfig = {
  low: {
    className: "badge-priority-low",
    label: "Low",
    icon: null,
  },
  medium: {
    className: "badge-priority-medium",
    label: "Medium",
    icon: null,
  },
  high: {
    className: "badge-priority-high",
    label: "High",
    icon: Flame,
  },
};

const statusConfig = {
  draft: { icon: Circle, color: "text-muted-foreground", label: "Draft" },
  active: { icon: Clock, color: "text-chart-2", label: "Active" },
  completed: {
    icon: CheckCircle2,
    color: "text-accent",
    label: "Completed",
  },
  archived: {
    icon: Archive,
    color: "text-muted-foreground",
    label: "Archived",
  },
};

export function TodoCard({ todo, compact = false }: TodoCardProps) {
  const [showDeleteDialog, setShowDeleteDialog] = useState(false);
  const updateTodo = useUpdateTodo();
  const deleteTodo = useDeleteTodo();

  const StatusIcon = statusConfig[todo.status].icon;

  const handleStatusToggle = async () => {
    const newStatus = todo.status === "completed" ? "active" : "completed";
    try {
      await updateTodo.mutateAsync({
        todoId: todo.id,
        body: {
          status: newStatus,
        },
      });
      toast.success(
        newStatus === "completed"
          ? "Task marked as completed!"
          : "Task marked as active",
      );
    } catch {
      toast.error("Failed to update task status");
    }
  };

  const handleDelete = async () => {
    try {
      await deleteTodo.mutateAsync({ todoId: todo.id });
      toast.success("Task deleted successfully");
      setShowDeleteDialog(false);
    } catch {
      toast.error("Failed to delete task");
    }
  };

  const isOverdue =
    todo.dueDate &&
    new Date(todo.dueDate) < new Date() &&
    todo.status !== "completed";

  return (
    <>
      <Card
        className={cn(
          "transition-all duration-300 shadow-soft hover-lift group",
          todo.status === "completed" && "opacity-80",
          compact
            ? "border-l-2 border-l-transparent hover:border-l-accent"
            : "hover:border-accent/30",
        )}
      >
        <CardContent className={cn("p-4", compact && "p-3")}>
          <div className="space-y-3">
            {/* Header with status and actions */}
            <div className="flex items-start justify-between gap-3">
              <div className="flex items-start gap-3 flex-1 min-w-0">
                <motion.div whileTap={{ scale: 0.9 }} className="mt-0.5">
                  <Checkbox
                    checked={todo.status === "completed"}
                    onCheckedChange={handleStatusToggle}
                    disabled={updateTodo.isPending}
                    className="data-[state=checked]:bg-accent data-[state=checked]:border-accent"
                  />
                </motion.div>
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2 mb-1">
                    <StatusIcon
                      className={cn("w-4 h-4", statusConfig[todo.status].color)}
                    />
                    <h3
                      className={cn(
                        "font-semibold truncate",
                        compact ? "text-sm" : "text-base",
                        todo.status === "completed" &&
                          "line-through text-muted-foreground",
                      )}
                    >
                      {todo.title}
                    </h3>
                    {isOverdue && (
                      <AlertTriangle className="w-4 h-4 text-destructive animate-pulse" />
                    )}
                  </div>

                  {todo.description && !compact && (
                    <p className="text-sm text-muted-foreground line-clamp-2 mb-2 leading-relaxed">
                      {todo.description}
                    </p>
                  )}
                </div>
              </div>

              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <Button
                    variant="ghost"
                    size="icon"
                    className="h-8 w-8 opacity-0 group-hover:opacity-100 transition-opacity"
                  >
                    <MoreHorizontal className="w-4 h-4" />
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end" className="w-48">
                  <TodoEditForm todo={todo}>
                    <DropdownMenuItem onSelect={(e) => e.preventDefault()}>
                      <Edit className="w-4 h-4 mr-2" />
                      Edit
                    </DropdownMenuItem>
                  </TodoEditForm>

                  <TodoCommentsDialog todoId={todo.id}>
                    <DropdownMenuItem onSelect={(e) => e.preventDefault()}>
                      <MessageSquare className="w-4 h-4 mr-2" />
                      Comments ({todo.comments?.length || 0})
                    </DropdownMenuItem>
                  </TodoCommentsDialog>

                  <DropdownMenuSeparator />

                  <DropdownMenuItem
                    onSelect={() => setShowDeleteDialog(true)}
                    className="text-destructive focus:text-destructive"
                  >
                    <Trash2 className="w-4 h-4 mr-2" />
                    Delete
                  </DropdownMenuItem>
                </DropdownMenuContent>
              </DropdownMenu>
            </div>

            {/* Metadata */}
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2 flex-wrap">
                {(() => {
                  const PriorityIcon = priorityConfig[todo.priority].icon;
                  return (
                    <span
                      className={cn(
                        "inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-medium",
                        priorityConfig[todo.priority].className,
                      )}
                    >
                      {PriorityIcon && <PriorityIcon className="w-3 h-3" />}
                      {priorityConfig[todo.priority].label}
                    </span>
                  );
                })()}

                {todo.category && (
                  <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-medium bg-muted/50 text-muted-foreground">
                    <span
                      className="w-2 h-2 rounded-full shadow-sm"
                      style={{ backgroundColor: todo.category.color }}
                    />
                    {todo.category.name}
                  </span>
                )}

                {todo.dueDate && (
                  <span
                    className={cn(
                      "inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-medium",
                      isOverdue
                        ? "bg-destructive/10 text-destructive border border-destructive/20"
                        : "bg-muted/50 text-muted-foreground",
                    )}
                  >
                    <Calendar className="w-3 h-3" />
                    {new Date(todo.dueDate).toLocaleDateString()}
                  </span>
                )}

                {todo.comments && todo.comments.length > 0 && (
                  <span className="inline-flex items-center gap-1 px-2 py-1 rounded-full text-xs bg-muted/30 text-muted-foreground">
                    <MessageSquare className="w-3 h-3" />
                    {todo.comments.length}
                  </span>
                )}

                {todo.attachments && todo.attachments.length > 0 && (
                  <span className="inline-flex items-center gap-1 px-2 py-1 rounded-full text-xs bg-muted/30 text-muted-foreground">
                    <Paperclip className="w-3 h-3" />
                    {todo.attachments.length}
                  </span>
                )}
              </div>

              {!compact && (
                <span className="text-xs text-muted-foreground/70">
                  {todo.status === "completed" && todo.completedAt
                    ? `Completed ${new Date(todo.completedAt).toLocaleDateString()}`
                    : `Updated ${new Date(todo.updatedAt).toLocaleDateString()}`}
                </span>
              )}
            </div>
          </div>
        </CardContent>
      </Card>

      <AlertDialog open={showDeleteDialog} onOpenChange={setShowDeleteDialog}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete Task</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to delete "{todo.title}"? This action cannot
              be undone.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction
              onClick={handleDelete}
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
              disabled={deleteTodo.isPending}
            >
              {deleteTodo.isPending ? "Deleting..." : "Delete"}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  );
}
