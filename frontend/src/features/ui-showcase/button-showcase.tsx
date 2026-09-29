"use client";

import { useState } from "react";
import { ArrowRight, Plus } from "lucide-react";

import { ContentSection } from "@/components/layout/content-section";
import { Button, type ButtonSize, type ButtonVariant } from "@/components/ui/button";

const buttonVariants: ButtonVariant[] = ["primary", "secondary", "ghost", "danger", "buy", "sell"];
const buttonSizes: ButtonSize[] = ["sm", "md", "lg"];

export function ButtonShowcase() {
  const [isSubmitting, setIsSubmitting] = useState(false);

  function simulateSubmit() {
    setIsSubmitting(true);
    window.setTimeout(() => setIsSubmitting(false), 1500);
  }

  return (
    <ContentSection title="Button" description="Variants, sizes, icons, loading and disabled states.">
      <div className="space-y-5">
        <div className="flex flex-wrap gap-3">
          {buttonVariants.map((variant) => (
            <Button key={variant} variant={variant} className="capitalize">
              {variant}
            </Button>
          ))}
        </div>
        <div className="flex flex-wrap items-center gap-3">
          {buttonSizes.map((size) => (
            <Button key={size} size={size} variant="secondary">
              Size {size}
            </Button>
          ))}
        </div>
        <div className="flex flex-wrap items-center gap-3">
          <Button>
            <Plus className="size-4" />
            New auction
          </Button>
          <Button variant="secondary">
            View market
            <ArrowRight className="size-4" />
          </Button>
          <Button isLoading={isSubmitting} onClick={simulateSubmit}>
            {isSubmitting ? "Placing order" : "Click to load"}
          </Button>
          <Button disabled>Disabled</Button>
        </div>
      </div>
    </ContentSection>
  );
}
