import { redirect } from "next/navigation";

interface Props {
  params: { orgName: string };
}

export default function OrgPage({ params }: Props) {
  redirect(`/org/${params.orgName}/workspace`);
}
