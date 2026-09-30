import { http } from "msw";

import type { CompletedUpgrade, CraftJob } from "@/features/workshop/api/workshop-api";
import { buildApiUrl } from "@/lib/api/api-config";
import { mockRecipes, mockUpgrades } from "@/mocks/fixtures/catalog-fixtures";
import { mockDataResponse, mockErrorResponse, mockListResponse, simulateLatency } from "@/mocks/utils/mock-responses";

const mockCraftSpeedup = 10;
const mockCraftJobs: CraftJob[] = [];

function deliverDueJobs() {
  for (const [jobIndex, craftJob] of mockCraftJobs.entries()) {
    if (craftJob.status === "CRAFTING" && new Date(craftJob.ends_at).getTime() <= Date.now()) {
      mockCraftJobs[jobIndex] = { ...craftJob, status: "DELIVERED", delivered_at: new Date().toISOString() };
    }
  }
}

export const workshopHandlers = [
  http.post(buildApiUrl("/crafts"), async ({ request }) => {
    await simulateLatency();
    deliverDueJobs();
    const { recipe_id: recipeID, quantity } = (await request.json()) as { recipe_id: string; quantity: number };
    const recipe = mockRecipes.find((candidate) => candidate.id === recipeID);
    if (!recipe) {
      return mockErrorResponse(422, "VALIDATION_FAILED", "Some fields are invalid", {
        recipe_id: "is not a known recipe",
      });
    }
    if (mockCraftJobs.some((craftJob) => craftJob.status === "CRAFTING")) {
      return mockErrorResponse(422, "LIMIT_EXCEEDED", "Your workshop is busy. Wait for the current craft to finish");
    }
    const startedAt = new Date();
    const craftJob: CraftJob = {
      id: crypto.randomUUID(),
      recipe_id: recipe.id,
      quantity,
      output_item_id: recipe.output_item_id,
      output_quantity: recipe.output_quantity * quantity,
      fee: (BigInt(recipe.fee) * BigInt(quantity)).toString(),
      inputs: recipe.inputs.map((input) => ({ item_id: input.item_id, quantity: input.quantity * quantity })),
      status: "CRAFTING",
      started_at: startedAt.toISOString(),
      ends_at: new Date(
        startedAt.getTime() + (recipe.craft_seconds * quantity * 1000) / mockCraftSpeedup,
      ).toISOString(),
      delivered_at: null,
    };
    mockCraftJobs.unshift(craftJob);
    return mockDataResponse(craftJob, 201);
  }),
  http.get(buildApiUrl("/crafts"), async ({ request }) => {
    await simulateLatency(150);
    deliverDueJobs();
    const requestUrl = new URL(request.url);
    const status = requestUrl.searchParams.get("status");
    return mockListResponse(
      status ? mockCraftJobs.filter((craftJob) => craftJob.status === status) : mockCraftJobs,
      requestUrl,
    );
  }),
  ...(["crafts", "purchases"] as const).map((path) =>
    http.post(buildApiUrl(`/upgrades/:upgradeID/${path}`), async ({ params }) => {
      await simulateLatency();
      const upgrade = mockUpgrades.find((candidate) => candidate.id === params.upgradeID);
      if (!upgrade) {
        return mockErrorResponse(404, "NOT_FOUND", "This upgrade does not exist");
      }
      const completedUpgrade: CompletedUpgrade = {
        id: crypto.randomUUID(),
        upgrade_id: upgrade.id,
        path: path === "crafts" ? "craft" : "buy",
        from_item_id: upgrade.from_item_id,
        to_item_id: upgrade.to_item_id,
        paid: path === "crafts" ? upgrade.craft_fee : (upgrade.buy_price ?? "0"),
        journal_id: crypto.randomUUID(),
      };
      return mockDataResponse(completedUpgrade, 201);
    }),
  ),
];
