import { useAuth } from '@/auth';
import UploadFile from '@/components/UploadFile/UploadFile';
import { createFileRoute } from '@tanstack/react-router'

export const Route = createFileRoute('/_auth/welcome')({
  component: RouteComponent,
})

function RouteComponent() {
  const auth = useAuth();
  return (<>
    <div>Hello {auth.tokenId}</div>
    <UploadFile />
  </>
  )
}
