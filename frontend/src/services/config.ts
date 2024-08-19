import type Config from "@/model/config";
import axios from "axios";

export default async function getConfig(): Promise<Config> {
  const res = await axios.get<Config>("/api/config");
  return res.data || {};
}
