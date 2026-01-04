import { useMutation, useQueryClient } from "@tanstack/react-query";
import { loginRequest } from "../api/auth"

export function useLogin() {
    const qc = useQueryClient()
    return useMutation({
        mutationFn: loginRequest,
        onSuccess: (_, variables) => {
            // refresh the /me query.            
            qc.invalidateQueries({ queryKey: ["me"] })

        },
        retry: false,
    })
}