import type { FieldValues, Path, UseFormSetError } from "react-hook-form";

import { isApiError } from "@/lib/api/api-error";

type ServerFieldMapping<FormValues extends FieldValues> =
  readonly Path<FormValues>[] | Record<string, Path<FormValues>>;

function capitalizeFirstLetter(text: string): string {
  return text.charAt(0).toUpperCase() + text.slice(1);
}

function formFieldFor<FormValues extends FieldValues>(
  serverFieldName: string,
  fieldMapping: ServerFieldMapping<FormValues>,
): Path<FormValues> | undefined {
  if (Array.isArray(fieldMapping)) {
    return fieldMapping.find((candidateName) => candidateName === serverFieldName);
  }
  return (fieldMapping as Record<string, Path<FormValues>>)[serverFieldName];
}

export function applyServerFieldErrors<FormValues extends FieldValues>(
  error: unknown,
  fieldMapping: ServerFieldMapping<FormValues>,
  setError: UseFormSetError<FormValues>,
): boolean {
  if (!isApiError(error)) {
    return false;
  }

  let hasAppliedFieldError = false;
  for (const [serverFieldName, fieldMessage] of Object.entries(error.fieldErrors)) {
    const formFieldName = formFieldFor(serverFieldName, fieldMapping);
    if (formFieldName) {
      setError(formFieldName, { type: "server", message: capitalizeFirstLetter(fieldMessage) });
      hasAppliedFieldError = true;
    }
  }
  return hasAppliedFieldError;
}
