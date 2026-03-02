import { useEffect, useRef, useState } from "react";
import { Check, Pencil, Plus, Trash2, X } from "lucide-react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  type Category,
  createCategory,
  deleteCategory,
  getCategories,
  updateCategory,
} from "@/api/categories";

interface CategorySelectorProps {
  selectedGuids: string[];
  onChange: (guids: string[]) => void;
}

export default function CategorySelector({
  selectedGuids,
  onChange,
}: Readonly<CategorySelectorProps>) {
  const queryClient = useQueryClient();
  const containerRef = useRef<HTMLDivElement>(null);

  const [isOpen, setIsOpen] = useState(false);
  const [newName, setNewName] = useState("");
  const [newColor, setNewColor] = useState("#94a3b8");
  const [editingGuid, setEditingGuid] = useState<string | null>(null);
  const [editName, setEditName] = useState("");
  const [editColor, setEditColor] = useState("");

  const { data: categories = [] } = useQuery<Category[]>({
    queryKey: ["categories"],
    queryFn: getCategories,
  });

  // Close popover on outside click.
  useEffect(() => {
    function onOutside(e: MouseEvent) {
      if (
        containerRef.current &&
        !containerRef.current.contains(e.target as Node)
      ) {
        setIsOpen(false);
        setEditingGuid(null);
      }
    }
    if (isOpen) document.addEventListener("mousedown", onOutside);
    return () => document.removeEventListener("mousedown", onOutside);
  }, [isOpen]);

  const createMutation = useMutation({
    mutationFn: ({ name, color }: { name: string; color: string }) =>
      createCategory(name, color),
    onSuccess: (newCat) => {
      queryClient.invalidateQueries({ queryKey: ["categories"] });
      onChange([...selectedGuids, newCat.guid]);
      setNewName("");
    },
  });

  const updateMutation = useMutation({
    mutationFn: ({
      guid,
      name,
      color,
    }: {
      guid: string;
      name: string;
      color: string;
    }) => updateCategory(guid, name, color),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["categories"] });
      setEditingGuid(null);
    },
  });

  const deleteMutation = useMutation({
    mutationFn: (guid: string) => deleteCategory(guid),
    onSuccess: (_, guid) => {
      queryClient.invalidateQueries({ queryKey: ["categories"] });
      onChange(selectedGuids.filter((g) => g !== guid));
    },
  });

  const selectedCategories = categories.filter((c) =>
    selectedGuids.includes(c.guid),
  );
  const availableCategories = categories.filter(
    (c) => !selectedGuids.includes(c.guid),
  );

  function toggleCategory(guid: string) {
    if (selectedGuids.includes(guid)) {
      onChange(selectedGuids.filter((g) => g !== guid));
    } else {
      onChange([...selectedGuids, guid]);
    }
  }

  function startEdit(cat: Category) {
    setEditingGuid(cat.guid);
    setEditName(cat.name);
    setEditColor(cat.color || "#94a3b8");
  }

  function commitEdit(guid: string) {
    if (editName.trim()) {
      updateMutation.mutate({ guid, name: editName.trim(), color: editColor });
    }
  }

  return (
    <div ref={containerRef} className="relative">
      {/* Selected tags row */}
      <div className="flex flex-wrap gap-1.5 min-h-[28px] items-center">
        {selectedCategories.map((cat) => (
          <span
            key={cat.guid}
            className="inline-flex items-center gap-1 rounded-full px-2.5 py-0.5 text-xs font-medium text-white"
            style={{ backgroundColor: cat.color || "#94a3b8" }}
          >
            {cat.name}
            <button
              type="button"
              onClick={() => toggleCategory(cat.guid)}
              className="hover:opacity-75 focus:outline-none"
              aria-label={`Remove ${cat.name}`}
            >
              <X size={10} />
            </button>
          </span>
        ))}
        <button
          type="button"
          onClick={() => setIsOpen((v) => !v)}
          className="inline-flex items-center gap-1 text-xs text-sky-600 hover:text-sky-800"
        >
          <Plus size={12} />
          {selectedCategories.length === 0 ? "Add category" : "More"}
        </button>
      </div>

      {/* Dropdown popover */}
      {isOpen && (
        <div className="absolute z-20 mt-1 w-72 rounded-lg border border-gray-200 bg-white shadow-lg p-3 space-y-1.5">
          {/* Available categories to select */}
          {availableCategories.map((cat) =>
            editingGuid === cat.guid ? (
              <div key={cat.guid} className="flex items-center gap-2">
                <input
                  type="color"
                  value={editColor}
                  onChange={(e) => setEditColor(e.target.value)}
                  className="h-6 w-6 rounded cursor-pointer border-0 p-0"
                />
                <input
                  type="text"
                  value={editName}
                  onChange={(e) => setEditName(e.target.value)}
                  className="flex-1 rounded border border-gray-300 px-2 py-0.5 text-xs focus:outline-none focus:ring-1 focus:ring-sky-400"
                  onKeyDown={(e) => {
                    if (e.key === "Enter") commitEdit(cat.guid);
                    if (e.key === "Escape") setEditingGuid(null);
                  }}
                  // biome-ignore lint/a11y/noAutofocus: intentional focus when editing
                  autoFocus
                />
                <button
                  type="button"
                  onClick={() => commitEdit(cat.guid)}
                  aria-label="Save"
                >
                  <Check size={13} className="text-green-600" />
                </button>
                <button
                  type="button"
                  onClick={() => setEditingGuid(null)}
                  aria-label="Cancel"
                >
                  <X size={13} className="text-gray-400" />
                </button>
              </div>
            ) : (
              <div key={cat.guid} className="flex items-center gap-2 group">
                <button
                  type="button"
                  onClick={() => toggleCategory(cat.guid)}
                  className="flex-1 flex items-center gap-2 text-left text-sm hover:text-sky-600"
                >
                  <span
                    className="h-3 w-3 rounded-full shrink-0"
                    style={{ backgroundColor: cat.color || "#94a3b8" }}
                  />
                  {cat.name}
                </button>
                <button
                  type="button"
                  onClick={() => startEdit(cat)}
                  className="opacity-0 group-hover:opacity-100 p-0.5"
                  aria-label={`Edit ${cat.name}`}
                >
                  <Pencil size={12} className="text-gray-400" />
                </button>
                <button
                  type="button"
                  onClick={() => deleteMutation.mutate(cat.guid)}
                  className="opacity-0 group-hover:opacity-100 p-0.5"
                  aria-label={`Delete ${cat.name}`}
                >
                  <Trash2 size={12} className="text-red-400" />
                </button>
              </div>
            ),
          )}

          {availableCategories.length === 0 && categories.length > 0 && (
            <p className="text-xs text-gray-400 text-center py-1">
              All categories selected
            </p>
          )}

          {/* Inline add new category */}
          <div className="pt-2 border-t border-gray-100">
            <p className="text-xs text-gray-400 mb-1.5">New category</p>
            <div className="flex items-center gap-2">
              <input
                type="color"
                value={newColor}
                onChange={(e) => setNewColor(e.target.value)}
                className="h-6 w-6 rounded cursor-pointer border-0 p-0"
                aria-label="Category color"
              />
              <input
                type="text"
                value={newName}
                onChange={(e) => setNewName(e.target.value)}
                placeholder="Category name"
                className="flex-1 rounded border border-gray-300 px-2 py-1 text-xs focus:outline-none focus:ring-1 focus:ring-sky-400"
                onKeyDown={(e) => {
                  if (e.key === "Enter" && newName.trim()) {
                    createMutation.mutate({
                      name: newName.trim(),
                      color: newColor,
                    });
                  }
                }}
              />
              <button
                type="button"
                disabled={!newName.trim() || createMutation.isPending}
                onClick={() =>
                  createMutation.mutate({
                    name: newName.trim(),
                    color: newColor,
                  })
                }
                className="rounded-md bg-sky-500 px-2 py-1 text-xs text-white hover:bg-sky-600 disabled:opacity-50"
              >
                Add
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
