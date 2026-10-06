import { activationEmailHandlers } from "@/mocks/handlers/activation-email-handlers";
import { fightHubHandlers } from "@/mocks/handlers/fight-hub-handlers";
import { levelFightHandlers } from "@/mocks/handlers/level-fight-handlers";
import { passwordResetHandlers } from "@/mocks/handlers/password-reset-handlers";

export const mockRequestHandlers = [
  ...activationEmailHandlers,
  ...passwordResetHandlers,
  ...fightHubHandlers,
  ...levelFightHandlers,
];
