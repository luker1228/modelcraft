"use client";

import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import { getToken, removeToken, refreshAccessToken } from "@bff/auth/public";
import { Button } from "@web/components/ui/button";
import { Input } from "@web/components/ui/input";
import { Label } from "@web/components/ui/label";
import { Badge } from "@web/components/ui/badge";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@web/components/ui/card";
import { Alert, AlertDescription } from "@web/components/ui/alert";
import { Building2, ArrowLeft, Loader2, CheckCircle2, AlertCircle } from "lucide-react";

// Function to convert display name to slug (underscores only, no hyphens)
function generateSlug(displayName: string): string {
  return displayName
    .toLowerCase()
    .trim()
    .replace(/[^\w\s]/g, '') // Remove special characters (keep letters, digits, underscores, spaces)
    .replace(/\s+/g, '_') // Replace spaces with underscores
    .replace(/_+/g, '_') // Replace multiple underscores with single underscore
    .replace(/^_+|_+$/g, '') // Remove leading/trailing underscores
    .replace(/^[^a-z]+/, ''); // Remove leading non-letter characters
}

// Add random suffix to ensure uniqueness
function generateUniqueSlug(displayName: string): string {
  const baseSlug = generateSlug(displayName);
  const randomSuffix = Math.floor(100 + Math.random() * 900); // 3-digit number

  if (!baseSlug) {
    // If slug is empty, generate a random one
    const randomWords = ['tech', 'labs', 'studio', 'works', 'systems', 'digital'];
    const randomWord = randomWords[Math.floor(Math.random() * randomWords.length)];
    return `${randomWord}_${randomSuffix}`;
  }

  // Truncate base to ensure total length stays within 24 chars
  const maxBaseLength = 20; // 20 + '_' + 3 digits = 24
  const truncatedBase = baseSlug.slice(0, maxBaseLength).replace(/_+$/, '');
  return `${truncatedBase}_${randomSuffix}`;
}

export default function CreateOrgPage() {
  const router = useRouter();
  const [displayName, setDisplayName] = useState("");
  const [generatedSlug, setGeneratedSlug] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState(false);

  // Auto-generate slug when display name changes
  useEffect(() => {
    if (displayName.trim()) {
      const slug = generateUniqueSlug(displayName);
      setGeneratedSlug(slug);
    } else {
      setGeneratedSlug("");
    }
  }, [displayName]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    // Validation
    if (!displayName.trim()) {
      setError("组织名称不能为空");
      return;
    }

    if (!generatedSlug) {
      setError("无法生成组织标识符");
      return;
    }

    setLoading(true);

    try {
      let token = getToken();

      if (!token) {
        token = await refreshAccessToken();
      }

      if (!token) {
        setError("会话已过期，请重新登录");
        setTimeout(() => router.push("/login"), 2000);
        return;
      }

      // Call backend API to initialize organization
      const response = await fetch(`${process.env.NEXT_PUBLIC_GATEWAY_URL ?? ''}/api/org/init`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify({
          displayName: displayName.trim(),
          organizationName: generatedSlug,
        }),
      });

      const data = await response.json() as Record<string, unknown>;

      if (!response.ok) {
        setError(typeof data.error === 'string' ? data.error : "初始化组织失败");
        setLoading(false);
        return;
      }

      // Success (alreadyExists=true means the user already had an org — still a success)
      setSuccess(true);
      console.log("[CreateOrg] Organization initialized:", data);

      // Redirect to root page after short delay
      setTimeout(() => {
        router.push("/");
      }, 1500);
    } catch (err) {
      console.error("[CreateOrg] Error:", err);
      setError(err instanceof Error ? err.message : "创建组织失败");
      setLoading(false);
    }
  };

  const handleBack = () => {
    router.push("/");
  };

  return (
    <div className="flex min-h-screen items-center justify-center bg-gradient-to-br from-slate-50 via-slate-100 to-slate-100 p-4">
      <div className="w-full max-w-md">
        {/* Back button */}
        <Button
          variant="ghost"
          size="sm"
          onClick={handleBack}
          className="mb-4"
        >
          <ArrowLeft className="mr-2 size-4" />
          返回组织列表
        </Button>

        <Card className="border-0 shadow-xl">
          <CardHeader className="text-center">
            <div className="mx-auto mb-4 inline-flex size-14 items-center justify-center rounded-2xl bg-primary text-primary-foreground">
              <Building2 className="size-7 text-white" />
            </div>
            <CardTitle className="text-2xl">创建组织</CardTitle>
            <CardDescription>
              为您的团队和项目创建一个新的组织空间
            </CardDescription>
          </CardHeader>

          <CardContent>
            {success ? (
              <Alert className="border-emerald-200 bg-emerald-50">
                <CheckCircle2 className="size-4 text-emerald-600" />
                <AlertDescription className="text-emerald-800">
                  组织创建成功！正在跳转...
                </AlertDescription>
              </Alert>
            ) : (
              <form onSubmit={handleSubmit} className="space-y-5">
                {/* Display Name */}
                <div className="space-y-2">
                  <Label htmlFor="displayName">
                    组织名称 <span className="text-destructive">*</span>
                  </Label>
                  <Input
                    id="displayName"
                    type="text"
                    placeholder="例如：我的科技公司"
                    value={displayName}
                    onChange={(e) => setDisplayName(e.target.value)}
                    disabled={loading}
                    className="text-base"
                    autoFocus
                  />
                  <p className="text-xs text-muted-foreground">
                    输入您的组织显示名称，可以使用中文、英文或其他字符
                  </p>
                </div>

                {/* Auto-generated Slug Preview */}
                {generatedSlug && (
                  <div className="space-y-2">
                    <Label className="text-muted-foreground">组织标识符（自动生成）</Label>
                    <div className="rounded-md border border-slate-200 bg-slate-50 p-3">
                      <div className="flex items-center gap-2">
                        <code className="flex-1 font-mono text-sm text-foreground">
                          {generatedSlug}
                        </code>
                        <Badge variant="secondary" className="text-xs">
                          自动
                        </Badge>
                      </div>
                      <p className="mt-2 text-xs text-muted-foreground">
                        URL: <span className="font-mono">modelcraft.com/org/{generatedSlug}</span>
                      </p>
                    </div>
                  </div>
                )}

                {/* Error Alert */}
                {error && (
                  <Alert variant="destructive">
                    <AlertCircle className="size-4" />
                    <AlertDescription>{error}</AlertDescription>
                  </Alert>
                )}

                {/* Submit Button */}
                <Button
                  type="submit"
                  className="w-full bg-primary hover:bg-primary/90"
                  disabled={loading || !displayName.trim()}
                >
                  {loading ? (
                    <>
                      <Loader2 className="mr-2 size-4 animate-spin" />
                      创建中...
                    </>
                  ) : (
                    "创建组织"
                  )}
                </Button>

                <p className="text-center text-xs text-muted-foreground">
                  创建组织后，您将成为该组织的所有者，可以邀请团队成员加入
                </p>
              </form>
            )}
          </CardContent>
        </Card>

        {/* Sign out link */}
        <div className="mt-6 text-center">
          <button
            onClick={() => {
              removeToken();
              localStorage.removeItem("defaultUserName");
              localStorage.removeItem("defaultOrgName");
              router.push("/login");
            }}
            className="text-sm text-muted-foreground transition-colors hover:text-foreground"
          >
            退出登录
          </button>
        </div>
      </div>
    </div>
  );
}
