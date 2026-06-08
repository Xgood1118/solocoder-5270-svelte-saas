import type { User, Org } from '$types';

declare global {
  namespace App {
    interface Locals {
      authToken: string | null;
      currentOrgId: string | null;
      user?: User;
      org?: Org;
    }
    interface PageData {
      user?: User | null;
      org?: Org | null;
    }
    interface Error {
      message: string;
      code?: string;
    }
  }
}

export {};
