"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation } from "@tanstack/react-query";
import { useForm } from "react-hook-form";
import { z } from "zod";

import { ContentSection } from "@/components/layout/content-section";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { useToast } from "@/components/ui/toast/use-toast";
import { useAuth } from "@/features/auth/session/use-auth";
import { usernamePattern } from "@/features/auth/signup/signup-schema";
import { updateUsername } from "@/features/settings/api/settings-api";
import { applyServerFieldErrors } from "@/utils/forms/apply-server-field-errors";

const profileSchema = z.object({
  username: z.string().regex(usernamePattern, "Use 3 to 20 letters, numbers or underscores"),
});

type ProfileFormValues = z.infer<typeof profileSchema>;

export function ProfileSection() {
  const { user, replaceUser } = useAuth();
  const { showToast } = useToast();

  const {
    register,
    handleSubmit,
    setError,
    reset,
    formState: { errors, isDirty },
  } = useForm<ProfileFormValues>({
    resolver: zodResolver(profileSchema),
    values: { username: user?.username ?? "" },
    mode: "onTouched",
  });

  const updateMutation = useMutation({
    mutationFn: (formValues: ProfileFormValues) => updateUsername(formValues.username.trim()),
    onSuccess: (updatedProfile) => {
      replaceUser(updatedProfile);
      reset({ username: updatedProfile.username });
      showToast({ tone: "success", title: "Username updated" });
    },
    onError: (error) => applyServerFieldErrors(error, ["username"], setError),
  });

  if (!user) {
    return null;
  }

  return (
    <ContentSection
      title="Profile"
      description="Your username is visible to other players on the exchange and in auctions."
    >
      <form
        onSubmit={handleSubmit((formValues) => updateMutation.mutate(formValues))}
        noValidate
        className="grid gap-4 sm:grid-cols-2"
      >
        <Input label="Email" value={user.email} readOnly disabled hint="Email changes aren't available yet." />
        <Input label="Username" errorMessage={errors.username?.message} {...register("username")} />
        <div className="sm:col-span-2">
          <Button type="submit" isLoading={updateMutation.isPending} disabled={!isDirty}>
            Save username
          </Button>
        </div>
      </form>
    </ContentSection>
  );
}
