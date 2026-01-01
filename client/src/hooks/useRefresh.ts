// hooks/useRefresh.ts
import { useMutation } from "@tanstack/react-query";
import { refreshRequest } from "../api/auth";

export function useRefresh() {
    return useMutation({
        mutationFn: refreshRequest,
    });
}