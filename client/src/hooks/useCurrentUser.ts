// hooks/useCurrentUser.ts
import { useQuery } from "@tanstack/react-query";
import { getMe } from "../api/auth";

export function useCurrentUser() {
    return useQuery({
        queryKey: ["me"],
        queryFn: getMe,
        retry: false, // don't retry 401
    });
}