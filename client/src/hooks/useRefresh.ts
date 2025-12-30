// hooks/useRefresh.ts
import { useMutation } from "@tanstack/react-query";
import { refreshRequest } from "../auth";

export function useRefresh() {
    return useMutation({
        mutationFn: refreshRequest,
    });
}