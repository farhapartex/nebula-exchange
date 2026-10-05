import { activationHandlers } from "@/mocks/handlers/activation-handlers";
import { authHandlers } from "@/mocks/handlers/auth-handlers";
import { fightHubHandlers } from "@/mocks/handlers/fight-hub-handlers";
import { levelFightHandlers } from "@/mocks/handlers/level-fight-handlers";
import { levelIntroHandlers } from "@/mocks/handlers/level-intro-handlers";
import { passwordResetHandlers } from "@/mocks/handlers/password-reset-handlers";
import { sessionHandlers } from "@/mocks/handlers/session-handlers";

export const mockRequestHandlers = [
  ...authHandlers,
  ...activationHandlers,
  ...passwordResetHandlers,
  ...sessionHandlers,
  ...fightHubHandlers,
  ...levelIntroHandlers,
  ...levelFightHandlers,
];
