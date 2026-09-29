import type { FieldValues, Path, UseFormSetError } from "react-hook-form";

import { isApiError } from "@/lib/api/api-error";

function capitalizeFirstLetter(text: string): string {
  return text.charAt(0).toUpperCase() + text.slice(1);
}

export function applyServerFieldErrors<FormValues extends FieldValues>(
  error: unknown,
  knownFieldNames: readonly Path<FormValues>[],
  setError: UseFormSetError<FormValues>,
): boolean {
  if (!isApiError(error)) {
    return false;
  }

  let hasAppliedFieldError = false;
  for (const [fieldName, fieldMessage] of Object.entries(error.fieldErrors)) {
    const knownFieldName = knownFieldNames.find((candidateName) => candidateName === fieldName);
    if (knownFieldName) {
      setError(knownFieldName, { type: "server", message: capitalizeFirstLetter(fieldMessage) });
      hasAppliedFieldError = true;
    }
  }
  return hasAppliedFieldError;
}
