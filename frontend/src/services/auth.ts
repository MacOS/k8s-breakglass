import type Config from "@/model/config";
import { UserManager, WebStorageStateStore, User, type UserManagerSettings, Log } from "oidc-client-ts";
import { ref } from "vue";

Log.setLogger(console);

export const AuthRedirect = "/auth/callback";

export interface State {
  path: string;
}

const user = ref<User>();

export default class AuthService {
  public userManager: UserManager;

  constructor(config: Config) {
    const baseURL = `${window.location.protocol}//${window.location.host}`;
    const settings: UserManagerSettings = {
      userStore: new WebStorageStateStore({ store: window.localStorage }),
      authority: config.oidcAuthority,
      client_id: config.oidcClientID,
      redirect_uri: baseURL + AuthRedirect,
      response_type: "code",
      scope: "openid",
      post_logout_redirect_uri: baseURL,
      filterProtocolClaims: true,
      automaticSilentRenew: true,
      accessTokenExpiringNotificationTimeInSeconds: 10,
    };

    this.userManager = new UserManager(settings);

    this.userManager.events.addUserLoaded((loadedUser) => {
      user.value = loadedUser;
    });
    this.userManager.getUser().then((u) => {
      if (u) {
        user.value = u;
      }
    });
  }

  public async getUser(): Promise<User | null> {
    return this.userManager.getUser();
  }

  public login(state?: State): Promise<void> {
    return this.userManager.signinRedirect({ state });
  }

  public logout(): Promise<void> {
    return this.userManager.signoutRedirect();
  }

  public async getAccessToken(): Promise<string> {
    const data = await this.userManager.getUser();
    return data?.access_token || "";
  }
}

export function useUser() {
  return user;
}
