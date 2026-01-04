export function enforceUserMatch({ auth, params }: any) {
    if (auth.isLoading) return;

    const authUser = auth.user?.username;
    const urlUser = params.user;

    if (authUser && urlUser && authUser !== urlUser) {
        return {
            redirect: {
                to: '/_auth/welcome/$user',
                params: { user: authUser },
            },
        };
    }
}

// import { createFileRoute, redirect } from '@tanstack/react-router'
// import { enforceUserMatch } from '../router/guards'

// export const Route = createFileRoute('/_auth/welcome/$user')({
//   beforeLoad: ({ context, params }) => {
//     const result = enforceUserMatch({ auth: context.auth, params });

//     if (result?.redirect) {
//       throw redirect(result.redirect);
//     }
//   },

//   // ⭐ Force beforeLoad to re-run when $user changes
//   shouldReload: ({ prev, next }) => prev.params.user !== next.params.user,

//   component: UserHomePage,
// });