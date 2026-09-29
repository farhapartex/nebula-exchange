import { ContentSection } from "@/components/layout/content-section";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardFooter, CardHeader } from "@/components/ui/card";
import { StatusBadge } from "@/components/ui/status-badge";

export function CardShowcase() {
  return (
    <ContentSection title="Card" description="Header with title, description and action, content, and footer.">
      <Card className="max-w-md">
        <CardHeader
          title="Asteroid Belt run"
          description="Scout · Drill T1 · 2 Fuel Cells"
          action={<StatusBadge label="Running" tone="info" isPulsing />}
        />
        <CardContent className="space-y-1 font-mono text-sm tabular-nums">
          <p className="text-foreground">Iron Ore 8–14</p>
          <p className="text-muted">Copper Ore 3–6</p>
        </CardContent>
        <CardFooter>
          <Button variant="ghost" size="sm">
            Abort
          </Button>
          <Button size="sm" disabled>
            Collect
          </Button>
        </CardFooter>
      </Card>
    </ContentSection>
  );
}
