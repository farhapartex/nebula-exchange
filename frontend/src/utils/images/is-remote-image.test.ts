import { describe, expect, it } from "vitest";

import { isRemoteImage } from "@/utils/images/is-remote-image";

describe("isRemoteImage", () => {
  it("recognises absolute links such as presigned storage urls", () => {
    expect(isRemoteImage("http://localhost:9000/street-born-assets/story/1-1/05-fire.webp?X-Amz-Signature=abc")).toBe(
      true,
    );
    expect(isRemoteImage("https://assets.streetborn.test/a.webp")).toBe(true);
  });

  it("treats files from the public folder as local", () => {
    expect(isRemoteImage("/story/1-1/05-fire.webp")).toBe(false);
  });
});
