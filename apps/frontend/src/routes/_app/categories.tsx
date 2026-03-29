import {
  useGetAllCategories,
  useDeleteCategory,
} from "@/api/hooks/use-category-query";
import { useDebounce } from "@/api/hooks/use-debounce";
import { CategoryCreateForm } from "@/components/categories/category-create-form";
import { CategoryEditForm } from "@/components/categories/category-edit-form";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from "@/components/ui/alert-dialog";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Input } from "@/components/ui/input";
import { Skeleton } from "@/components/ui/skeleton";
import { createFileRoute } from "@tanstack/react-router";
import {
  Plus,
  Search,
  MoreHorizontal,
  Edit,
  Trash2,
  Folder,
  Sparkles,
  X,
  ChevronLeft,
  ChevronRight,
} from "lucide-react";
import { motion, AnimatePresence } from "motion/react";
import { useState } from "react";
import { toast } from "sonner";

export const Route = createFileRoute("/_app/categories")({
  component: CategoriesPage,
});

function CategoriesPage() {
  const [searchQuery, setSearchQuery] = useState("");
  const [page, setPage] = useState(1);

  const debouncedSearch = useDebounce(searchQuery, 300);

  const { data: categories, isLoading } = useGetAllCategories({
    query: {
      page,
      limit: 20,
      search: debouncedSearch || undefined,
      sort: "name",
      order: "asc",
    },
  });

  const deleteCategory = useDeleteCategory();

  const handleDeleteCategory = async (
    categoryId: string,
    categoryName: string,
  ) => {
    try {
      await deleteCategory.mutateAsync({ categoryId });
      toast.success(`Category "${categoryName}" deleted successfully`);
    } catch {
      toast.error("Failed to delete category");
    }
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <motion.div
          initial={{ opacity: 0, y: 10 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.4 }}
        >
          <h1 className="text-3xl md:text-4xl font-bold tracking-tight">
            Categories
          </h1>
          <p className="text-muted-foreground mt-1">
            Organize your tasks with custom categories
          </p>
        </motion.div>
        <motion.div
          initial={{ opacity: 0, scale: 0.95 }}
          animate={{ opacity: 1, scale: 1 }}
          transition={{ duration: 0.4, delay: 0.1 }}
        >
          <CategoryCreateForm>
            <Button className="shadow-soft press-scale">
              <Plus className="w-4 h-4 mr-2" />
              New Category
            </Button>
          </CategoryCreateForm>
        </motion.div>
      </div>

      {/* Search */}
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.4, delay: 0.1 }}
      >
        <Card className="shadow-soft">
          <CardContent className="p-4">
            <div className="relative">
              <Search className="absolute left-3.5 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
              <Input
                placeholder="Search categories..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="pl-10 h-11 bg-muted/30 border-transparent focus:border-accent focus:bg-background transition-colors"
              />
              {searchQuery && (
                <button
                  onClick={() => setSearchQuery("")}
                  className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground transition-colors"
                >
                  <X className="w-4 h-4" />
                </button>
              )}
            </div>
          </CardContent>
        </Card>
      </motion.div>

      {/* Categories Grid */}
      <div className="space-y-4">
        {isLoading ? (
          <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
            {Array.from({ length: 8 }).map((_, i) => (
              <Card key={i} className="shadow-soft">
                <CardContent className="p-5">
                  <div className="space-y-4">
                    <div className="flex items-center justify-between">
                      <Skeleton className="h-5 w-5 rounded-lg" />
                      <Skeleton className="h-6 w-6 rounded-lg" />
                    </div>
                    <Skeleton className="h-5 w-2/3" />
                    <Skeleton className="h-4 w-full" />
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>
        ) : categories?.data?.length ? (
          <motion.div
            className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4"
            initial="hidden"
            animate="visible"
          >
            {categories.data.map((category, index) => (
              <motion.div
                key={category.id}
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ duration: 0.4, delay: index * 0.05 }}
              >
                <Card className="group relative shadow-soft hover-lift transition-all duration-300 hover:border-accent/30 overflow-hidden">
                  {/* Accent stripe */}
                  <div
                    className="absolute top-0 left-0 right-0 h-1"
                    style={{ backgroundColor: category.color }}
                  />
                  <CardContent className="p-5 pt-6">
                    <div className="space-y-4">
                      <div className="flex items-center justify-between">
                        <div
                          className="w-10 h-10 rounded-xl flex items-center justify-center shadow-sm"
                          style={{
                            backgroundColor: `${category.color}20`,
                          }}
                        >
                          <Folder
                            className="w-5 h-5"
                            style={{ color: category.color }}
                          />
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
                          <DropdownMenuContent align="end">
                            <CategoryEditForm category={category}>
                              <DropdownMenuItem
                                onSelect={(e) => e.preventDefault()}
                              >
                                <Edit className="w-4 h-4 mr-2" />
                                Edit
                              </DropdownMenuItem>
                            </CategoryEditForm>
                            <AlertDialog>
                              <AlertDialogTrigger asChild>
                                <DropdownMenuItem
                                  onSelect={(e) => e.preventDefault()}
                                  className="text-destructive focus:text-destructive"
                                >
                                  <Trash2 className="w-4 h-4 mr-2" />
                                  Delete
                                </DropdownMenuItem>
                              </AlertDialogTrigger>
                              <AlertDialogContent>
                                <AlertDialogHeader>
                                  <AlertDialogTitle>
                                    Delete category?
                                  </AlertDialogTitle>
                                  <AlertDialogDescription>
                                    This will permanently delete "
                                    {category.name}". Tasks with this category
                                    will become uncategorized.
                                  </AlertDialogDescription>
                                </AlertDialogHeader>
                                <AlertDialogFooter>
                                  <AlertDialogCancel>Cancel</AlertDialogCancel>
                                  <AlertDialogAction
                                    className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
                                    onClick={() =>
                                      handleDeleteCategory(
                                        category.id,
                                        category.name,
                                      )
                                    }
                                  >
                                    Delete
                                  </AlertDialogAction>
                                </AlertDialogFooter>
                              </AlertDialogContent>
                            </AlertDialog>
                          </DropdownMenuContent>
                        </DropdownMenu>
                      </div>
                      <div>
                        <h3 className="font-semibold text-lg">
                          {category.name}
                        </h3>
                        {category.description && (
                          <p className="text-sm text-muted-foreground mt-1 line-clamp-2">
                            {category.description}
                          </p>
                        )}
                      </div>
                      <div className="flex items-center justify-end pt-2 border-t border-border/50">
                        <div
                          className="w-3 h-3 rounded-full shadow-sm ring-2 ring-background"
                          style={{ backgroundColor: category.color }}
                        />
                      </div>
                    </div>
                  </CardContent>
                </Card>
              </motion.div>
            ))}
          </motion.div>
        ) : (
          <motion.div
            className="py-16 text-center"
            initial={{ opacity: 0, scale: 0.95 }}
            animate={{ opacity: 1, scale: 1 }}
            transition={{ duration: 0.4 }}
          >
            <div className="w-16 h-16 mx-auto rounded-2xl bg-muted/50 flex items-center justify-center mb-4">
              <Sparkles className="w-8 h-8 text-muted-foreground" />
            </div>
            <p className="text-lg font-medium text-muted-foreground">
              No categories found
            </p>
            <p className="text-sm text-muted-foreground/70 mt-1">
              {searchQuery
                ? "Try a different search term"
                : "Create your first category to organize tasks"}
            </p>
            {!searchQuery && (
              <CategoryCreateForm>
                <Button variant="outline" className="mt-4">
                  <Plus className="w-4 h-4 mr-2" />
                  Create Category
                </Button>
              </CategoryCreateForm>
            )}
          </motion.div>
        )}

        {/* Pagination */}
        {categories && categories.totalPages > 1 && (
          <motion.div
            className="flex items-center justify-between pt-4"
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ delay: 0.2 }}
          >
            <p className="text-sm text-muted-foreground">
              Page {page} of {categories.totalPages}
            </p>
            <div className="flex gap-2">
              <Button
                variant="outline"
                size="sm"
                disabled={page <= 1}
                onClick={() => setPage(page - 1)}
                className="gap-1"
              >
                <ChevronLeft className="w-4 h-4" />
                Previous
              </Button>
              <Button
                variant="outline"
                size="sm"
                disabled={page >= categories.totalPages}
                onClick={() => setPage(page + 1)}
                className="gap-1"
              >
                Next
                <ChevronRight className="w-4 h-4" />
              </Button>
            </div>
          </motion.div>
        )}
      </div>
    </div>
  );
}
