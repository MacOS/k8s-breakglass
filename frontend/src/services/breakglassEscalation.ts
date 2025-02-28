
import axios, { AxiosHeaders } from "axios";

import type AuthService from "@/services/auth";

export default class BreakglassEscalationService {
  private client = axios.create({
    baseURL: "/api/breakglassEscalation/",
  });
  private auth: AuthService;

  constructor(auth: AuthService) {
    this.auth = auth;

    this.client.interceptors.request.use(async (req) => {
      if (!req.headers) {
        req.headers = {} as AxiosHeaders;
      }
      req.headers["Authorization"] = `Bearer ${await this.auth.getAccessToken()}`;
      return req;
    });
  }


  public async getEscalations() {
    return await this.client.get("/escalations", {})
  }
}
