import { activationEmailHandlers } from "@/mocks/handlers/activation-email-handlers";
import { fightHubHandlers } from "@/mocks/handlers/fight-hub-handlers";
import { passwordResetHandlers } from "@/mocks/handlers/password-reset-handlers";
import { trainingFightHandlers } from "@/mocks/handlers/training-fight-handlers";

export const mockRequestHandlers = [
  ...activationEmailHandlers,
  ...passwordResetHandlers,
  ...fightHubHandlers,
  ...trainingFightHandlers,
];
