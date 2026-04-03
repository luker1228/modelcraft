"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import {
  redirectToLogin,
  getToken,
  isTokenExpired,
  getUserInfoFromToken,
  removeToken,
} from "@bff/auth/casdoor";
import { refreshAccessToken } from "@bff/auth/public";
import { Button } from "@web/components/ui/button";
import { Badge } from "@web/components/ui/badge";
import { Card, CardContent } from "@web/components/ui/card";
import { Separator } from "@web/components/ui/separator";
import { Sparkles, Shield, Globe, ShieldCheck, Link } from "lucide-react";

export default function LoginPage() {
  const router = useRouter();

  useEffect(() => {
    async function checkAlreadyLoggedIn() {
      let token = getToken();
      console.log("[LoginPage] Checking authentication...", { hasToken: !!token });

      // Try silent refresh in case user navigated to /login with valid cookie
      if (!token) {
        token = await refreshAccessToken();
      }

      if (token && !isTokenExpired(token)) {
        const userInfo = getUserInfoFromToken(token);
        console.log("[LoginPage] Already authenticated:", userInfo?.id, "→ redirecting");
        router.push("/org-selector");
      } else {
        if (token) removeToken();
        console.log("[LoginPage] No valid session, showing login page");
      }
    }

    checkAlreadyLoggedIn();
  }, [router]);

  const handleLogin = () => {
    console.log("[LoginPage] Login button clicked, redirecting to Casdoor...");
    redirectToLogin();
  };

  const features = [
    {
      icon: <Globe className="size-5" strokeWidth={1.5} />,
      title: "数据库 HTTP 接口",
      desc: "自动将数据库表暴露为 GraphQL API 接口",
      bgColor: "bg-[#2563eb]",
    },
    {
      icon: <ShieldCheck className="size-5" strokeWidth={1.5} />,
      title: "行级权限控制 (RLS)",
      desc: "基于用户身份的细粒度数据访问策略",
      bgColor: "bg-[#059669]",
    },
    {
      icon: <Link className="size-5" strokeWidth={1.5} />,
      title: "逻辑外键与枚举",
      desc: "无需数据库约束，灵活定义关联关系与枚举类型",
      bgColor: "bg-[#d97706]",
    },
  ];

  return (
    <div className="relative flex min-h-screen overflow-hidden bg-gray-50">
      {/* Clean background with subtle border grid */}

      {/* Left brand area */}
      <div className="relative z-10 hidden lg:flex lg:w-1/2">
        <div className="mx-auto flex w-full max-w-2xl flex-col justify-between p-12">
          {/* Logo */}
          <div className="flex items-center gap-3">
            <div className="flex size-12 items-center justify-center rounded-lg bg-[#2563eb] shadow-sm">
              <Sparkles className="size-6 text-white" strokeWidth={1.5} />
            </div>
            <span className="text-2xl font-semibold text-foreground">
              ModelCraft
            </span>
          </div>

          {/* Middle content */}
          <div className="space-y-8">
            <div className="space-y-6">
              <span className="inline-flex select-none rounded bg-[#dbeafe] px-4 py-2 text-sm font-semibold text-[#2563eb]">
                ✨ 数据库能力即服务平台
              </span>

              <h1 className="text-5xl font-semibold leading-tight text-foreground">
                让数据库能力触手可及
              </h1>

              <p className="max-w-lg text-lg leading-relaxed text-muted-foreground">
                为开发者打造的数据库接口平台。HTTP 接口、行级权限、逻辑关联 —— 无需编写后端代码。
              </p>
            </div>

            {/* Feature cards */}
            <div className="grid grid-cols-1 gap-4">
              {features.map((feature, i) => (
                <Card
                  key={i}
                  className="group cursor-default rounded-lg border border-gray-200 bg-white transition-all duration-200 hover:border-[#dbeafe] hover:bg-white hover:shadow-sm"
                >
                  <CardContent className="flex items-start gap-4 p-4">
                    <div
                      className={`flex size-10 flex-shrink-0 items-center justify-center rounded-md text-white ${feature.bgColor}`}
                    >
                      {feature.icon}
                    </div>
                    <div className="min-w-0 flex-1">
                      <h3 className="mb-1 font-semibold text-foreground">
                        {feature.title}
                      </h3>
                      <p className="text-sm text-muted-foreground">{feature.desc}</p>
                    </div>
                  </CardContent>
                </Card>
              ))}
            </div>
          </div>

          {/* Footer */}
          <div className="flex items-center justify-between text-sm text-muted-foreground">
            <span>&copy; 2024 ModelCraft. 保留所有权利</span>
            <div className="flex gap-4">
              <a href="#" className="transition-colors hover:text-foreground">
                隐私政策
              </a>
              <a href="#" className="transition-colors hover:text-foreground">
                服务条款
              </a>
            </div>
          </div>
        </div>
      </div>

      {/* Right login form */}
      <div className="relative z-10 flex flex-1 items-center justify-center p-6">
        <Card className="w-full max-w-md rounded-lg border border-gray-200 bg-white shadow-sm">
          <CardContent className="p-8 sm:p-10">
            {/* Mobile logo */}
            <div className="mb-8 flex items-center justify-center gap-3 lg:hidden">
              <div className="flex size-10 items-center justify-center rounded-lg bg-[#2563eb]">
                <Sparkles className="size-5 text-white" strokeWidth={1.5} />
              </div>
              <span className="text-xl font-semibold text-foreground">
                ModelCraft
              </span>
            </div>

            {/* Form title */}
            <div className="mb-8 text-center lg:text-left">
              <h2 className="mb-2 text-3xl font-semibold text-foreground">
                欢迎回来
              </h2>
              <p className="text-muted-foreground">登录以访问您的工作空间</p>
            </div>

            {/* Casdoor login button */}
            <Button
              onClick={handleLogin}
              className="h-11 w-full cursor-pointer rounded-md border-0 bg-[#2563eb] font-semibold text-white transition-all duration-200 hover:bg-[#1d4ed8]"
            >
              <Shield className="mr-2 size-4" strokeWidth={1.5} />
              使用 Casdoor 登录
            </Button>

            {/* Features list */}
            <div className="mt-8 space-y-4">
              <Separator className="bg-gray-200" />
              <p className="pt-2 text-xs font-semibold uppercase tracking-wide text-muted-foreground">
                平台特性
              </p>
              {features.map((feature, i) => (
                <div key={i} className="flex items-start gap-3">
                  <div
                    className={`flex size-8 flex-shrink-0 items-center justify-center rounded-md text-white ${feature.bgColor}`}
                  >
                    {feature.icon}
                  </div>
                  <div className="min-w-0 flex-1">
                    <p className="text-sm font-semibold text-foreground">
                      {feature.title}
                    </p>
                    <p className="text-xs font-normal text-muted-foreground">{feature.desc}</p>
                  </div>
                </div>
              ))}
            </div>

            {/* Footer */}
            <div className="mt-8 space-y-4">
              <Separator className="bg-gray-100" />
              <p className="pt-2 text-center text-xs leading-relaxed text-muted-foreground">
                登录即表示您同意我们的{" "}
                <a href="#" className="text-[#2563eb] hover:underline">
                  服务条款
                </a>{" "}
                和{" "}
                <a href="#" className="text-[#2563eb] hover:underline">
                  隐私政策
                </a>
              </p>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
