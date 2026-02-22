import { uploadFile, type UploadMetadata } from "@/api/upload";
import { useState, useRef } from "react";
import PurchaseInfo from "./PurchaseInfo/PurchaseInfo";

// Field configuration - add new fields here to extend the form
type FieldType = "text" | "textarea";

interface FieldConfig {
  name: keyof UploadMetadata;
  label: string;
  type: FieldType;
  required?: boolean;
}

const FORM_FIELDS: FieldConfig[] = [
  { name: "title", label: "Title", type: "text", required: true },
  { name: "description", label: "Description", type: "textarea" },
  // Add new fields here, e.g.:
  // { name: 'tags', label: 'Tags', type: 'text' },
  // { name: 'category', label: 'Category', type: 'text' },
];

const getInitialMetadata = (): UploadMetadata => ({
  ...FORM_FIELDS.reduce((acc, field) => ({ ...acc, [field.name]: "" }), {}),
  buyAt: "",
  buyFrom: "",
  buyPrice: "",
});

export default function UploadFile() {
  const [metadata, setMetadata] = useState<UploadMetadata>(getInitialMetadata);
  const [file, setFile] = useState<File | null>(null);
  const [success, setSuccess] = useState(false);

  const fileInputRef = useRef<HTMLInputElement | null>(null);

  const updateField = (name: string, value: string) => {
    setMetadata((prev) => ({ ...prev, [name]: value }));
  };

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    const ok = await uploadFile(metadata, file);

    // If upload sucecss, clear the form.
    if (ok) {
      setSuccess(true);
      handleClear();
      // Auto-hide after 3 seconds.
      setTimeout(() => setSuccess(false), 2000);
    }
  }

  // handle cancel.
  function handleClear() {
    setMetadata(getInitialMetadata());
    setFile(null);

    // Reset the file input visually.
    if (fileInputRef.current) {
      fileInputRef.current.value = "";
    }
  }

  return (
    <div>
      <div>
        {success && (
          <div className="mb-4 my-1 rounded-md bg-green-100 px-4 py-2 text-green-800 border border-green-300">
            Uploaded successfully
          </div>
        )}
        <form onSubmit={handleSubmit}>
          <div className="space-y-12">
            <div className="border-b border-gray-900/10 pb-12">
              <div className="mt-10 grid grid-cols-1 gap-x-6 gap-y-8 sm:grid-cols-6">
                {FORM_FIELDS.map((field) => (
                  <div key={field.name} className="col-span-full">
                    <label
                      htmlFor={field.name}
                      className="block text-sm/6 font-medium text-gray-900"
                    >
                      {field.label}
                      {field.required && (
                        <span className="text-red-500">*</span>
                      )}
                    </label>
                    <div className="mt-2">
                      {field.type === "text" ? (
                        <div className="flex items-center rounded-md bg-white pl-3 outline-1 -outline-offset-1 outline-gray-300 focus-within:outline-2 focus-within:-outline-offset-2 focus-within:outline-indigo-600">
                          <input
                            id={field.name}
                            name={field.name}
                            type="text"
                            value={metadata[field.name] ?? ""}
                            onChange={(e) =>
                              updateField(field.name, e.target.value)
                            }
                            className="block min-w-0 grow bg-white py-1.5 pr-3 pl-1 text-base text-gray-900 placeholder:text-gray-400 focus:outline-none sm:text-sm/6"
                          />
                        </div>
                      ) : (
                        <textarea
                          id={field.name}
                          name={field.name}
                          value={metadata[field.name] ?? ""}
                          onChange={(e) =>
                            updateField(field.name, e.target.value)
                          }
                          className="block w-full rounded-md bg-white px-3 py-1.5 text-base text-gray-900 outline-1 -outline-offset-1 outline-gray-300 placeholder:text-gray-400 focus:outline-2 focus:-outline-offset-2 focus:outline-indigo-600 sm:text-sm/6"
                        />
                      )}
                    </div>
                  </div>
                ))}

                <div className="col-span-full">
                  <label
                    htmlFor="upload"
                    className="block text-sm font-medium text-gray-900"
                  >
                    Upload Files
                  </label>

                  <div className="mt-2 flex justify-center rounded-lg border border-dashed border-gray-900/25 px-6 py-10">
                    <div className="text-center">
                      <svg
                        viewBox="0 0 24 24"
                        fill="currentColor"
                        className="mx-auto size-12 text-gray-300"
                      >
                        <path d="M1.5 6a2.25 2.25 0 0 1 2.25-2.25h16.5A2.25 2.25 0 0 1 22.5 6v12a2.25 2.25 0 0 1-2.25 2.25H3.75A2.25 2.25 0 0 1 1.5 18V6ZM3 16.06V18c0 .414.336.75.75.75h16.5A.75.75 0 0 0 21 18v-1.94l-2.69-2.689a1.5 1.5 0 0 0-2.12 0l-.88.879.97.97a.75.75 0 1 1-1.06 1.06l-5.16-5.159a1.5 1.5 0 0 0-2.12 0L3 16.061Zm10.125-7.81a1.125 1.125 0 1 1 2.25 0 1.125 1.125 0 0 1-2.25 0Z" />
                      </svg>

                      <div className="mt-4 flex text-sm text-gray-600">
                        <label
                          htmlFor="file"
                          className="relative cursor-pointer rounded-md font-semibold text-slate-900 hover:text-slate-500"
                        >
                          <span>Upload a file</span>
                          <input
                            id="file"
                            type="file"
                            ref={fileInputRef}
                            className="sr-only"
                            onChange={(e) =>
                              setFile(e.target.files?.[0] || null)
                            }
                          />
                        </label>
                        <p className="pl-1">or drag and drop</p>
                      </div>

                      <p className="text-xs text-gray-600">
                        PNG, JPG, GIF up to 5MB
                      </p>
                    </div>
                  </div>

                  {/* Show selected file */}
                  {file && (
                    <div className="mt-4 text-sm text-gray-700">
                      Selected file:{" "}
                      <span className="font-medium">{file.name}</span>
                    </div>
                  )}

                  {/* Optional image preview */}
                  {file && file.type.startsWith("image/") && (
                    <img
                      src={URL.createObjectURL(file)}
                      alt="Preview"
                      className="mt-4 h-32 object-contain rounded-md border"
                    />
                  )}
                </div>
              </div>
              <PurchaseInfo
                values={{
                  buyAt: metadata.buyAt ?? "",
                  buyFrom: metadata.buyFrom ?? "",
                  buyPrice: metadata.buyPrice ?? "",
                }}
                onChange={updateField}
              />
            </div>
          </div>

          <div className="mt-6 flex items-center justify-end gap-x-6">
            <button
              type="button"
              onClick={handleClear}
              className="text-sm font-semibold text-gray-900 px-3 py-2 rounded-md 
                    hover:bg-gray-200 
                    focus-visible:outline-2 
                    focus-visible:outline-offset-2 
                    focus-visible:outline-indigo-600"
            >
              Clear
            </button>
            <button
              type="submit"
              className="rounded-md bg-slate-900 px-3 py-2 text-sm font-semibold text-white shadow-xs 
                    hover:bg-slate-600 
                    focus-visible:outline-2 
                    focus-visible:outline-offset-2 
                    focus-visible:outline-indigo-600"
            >
              Save
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
