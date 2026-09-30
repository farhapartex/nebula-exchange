import type {
  CatalogItem,
  CatalogItemDetail,
  Recipe,
  ShopItem,
  Upgrade,
  Zone,
} from "@/features/catalog/api/catalog-types";
import { requestAllPages, requestData } from "@/lib/api/api-client";

export function fetchCatalogItems(): Promise<CatalogItem[]> {
  return requestAllPages<CatalogItem>("/items");
}

export function fetchCatalogItem(itemID: number): Promise<CatalogItemDetail> {
  return requestData<CatalogItemDetail>(`/items/${itemID}`);
}

export function fetchRecipes(): Promise<Recipe[]> {
  return requestAllPages<Recipe>("/recipes");
}

export function fetchUpgrades(): Promise<Upgrade[]> {
  return requestAllPages<Upgrade>("/upgrades");
}

export function fetchZones(): Promise<Zone[]> {
  return requestAllPages<Zone>("/zones");
}

export function fetchShopItems(): Promise<ShopItem[]> {
  return requestAllPages<ShopItem>("/shop-items");
}
