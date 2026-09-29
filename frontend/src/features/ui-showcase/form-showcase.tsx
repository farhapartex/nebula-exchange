"use client";

import { useState } from "react";

import { ContentSection } from "@/components/layout/content-section";
import { Input } from "@/components/ui/input";
import { Select, type SelectOption } from "@/components/ui/select";

const zoneOptions: SelectOption[] = [
  { value: "asteroid-belt", label: "Asteroid Belt" },
  { value: "crystal-moon", label: "Crystal Moon" },
  { value: "plasma-nebula", label: "Plasma Nebula" },
  { value: "deep-void", label: "Deep Void (locked)", isDisabled: true },
];

export function FormShowcase() {
  const [selectedZone, setSelectedZone] = useState<string>();

  return (
    <ContentSection title="Input and Select" description="Labels, hints, errors, adornments and disabled fields.">
      <div className="grid gap-5 sm:grid-cols-2">
        <Input label="Email" type="email" placeholder="pilot@nebula.test" hint="We send a 6-digit code here." />
        <Input label="Username" defaultValue="ab" errorMessage="Must be 3 to 20 characters." />
        <Input
          label="Price"
          inputMode="decimal"
          placeholder="0.0000"
          trailingAdornment="NC"
          className="font-mono tabular-nums"
        />
        <Input label="Wallet address" value="0x9f2c…a41e" disabled readOnly />
        <Select
          label="Mission zone"
          options={zoneOptions}
          value={selectedZone}
          onValueChange={setSelectedZone}
          placeholder="Choose a zone"
          hint={selectedZone ? `Selected: ${selectedZone}` : "Deep Void needs a T4 drill."}
        />
        <Select label="Ship" options={[]} placeholder="No ships available" errorMessage="Buy a Scout in the shop." />
      </div>
    </ContentSection>
  );
}
