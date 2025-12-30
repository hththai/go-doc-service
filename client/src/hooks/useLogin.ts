import { useMutation } from "@tanstack/react-query";
import { loginRequest } from "../auth"

export function useLogin() {
    return useMutation({
        mutationFn: loginRequest
    })
}