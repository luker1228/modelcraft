import { redirect } from 'next/navigation'

interface Props {
  params: { orgName: string; projectSlug: string }
}

export default function ProjectHomePage({ params }: Props) {
  redirect(`/org/${params.orgName}/project/${params.projectSlug}/model-editor`)
}
