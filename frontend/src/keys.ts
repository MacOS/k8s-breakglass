import type { InjectionKey } from "vue";
import type AuthService from "@/services/auth";

export const AuthKey: InjectionKey<AuthService> = Symbol("auth");
