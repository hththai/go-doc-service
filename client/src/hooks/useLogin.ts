import { useMutation, useQueryClient } from "@tanstack/react-query";
import { loginRequest } from "../auth"

export function useLogin() {
    const qc = useQueryClient()
    return useMutation({
        mutationFn: loginRequest,
        onSuccess: (_, variables) => {
            localStorage.setItem('usr', variables.username);

            // refresh the /me query.
            qc.invalidateQueries({ queryKey: ["me"] })

        }
    })
}