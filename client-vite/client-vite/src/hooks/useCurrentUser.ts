// hooks/useCurrentUser.ts
import { useQuery } from "@tanstack/react-query";
import { getMe } from "../api/auth";

export function useCurrentUser(tokenId?: string | null) {
    return useQuery({
        queryKey: ["me", tokenId],
        queryFn: () => getMe({ tokenId: tokenId! }),
        enabled: !!tokenId,
        retry: false, // don't retry 401
    });
};

