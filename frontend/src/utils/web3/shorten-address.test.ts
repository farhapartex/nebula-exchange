import { describe, expect, it } from "vitest";

import { shortenAddress } from "@/utils/web3/shorten-address";

describe("shortenAddress", () => {
  it("keeps the start and end of an address", () => {
    expect(shortenAddress("0x8ba1f109551bD432803012645Ac136ddd64DBA72")).toBe("0x8ba1…BA72");
  });

  it("leaves short values alone", () => {
    expect(shortenAddress("0x1234")).toBe("0x1234");
  });
});
