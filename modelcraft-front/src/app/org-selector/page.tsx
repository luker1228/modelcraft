"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import {
  getToken,
  getUserInfoFromToken,
  removeToken,
  refreshAccessToken,
} from "@bff/auth/public";
import { useOrganizationStore } from "@shared/stores/organization";
import { Button } from "@web/components/ui/button";
import { Input } from "@web/components/ui/input";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@web/components/ui/dropdown-menu";
import {
  Building2,
  Users,
  ArrowRight,
  Plus,
  LogOut,
  Boxes,
  Search,
  RefreshCw,
  ShieldCheck,
} from "lucide-react";
import { getSmartRedirectUrl } from "@web/routing/smart-redirect";

// MembershipInfo from API response
interface MembershipInfo {
  orgId: string;
  orgName: string;
  displayName: string;
  role: string;
  joinedAt: string;
}

// Function to generate slug from display name
function generateSlug(displayName: string): string {
  return displayName
    .toLowerCase()
    .trim()
    .replace(/[^\w\s-]/g, '')
    .replace(/\s+/g, '-')
    .replace(/-+/g, '-')
    .replace(/^-+|-+$/g, '');
}

// Generate unique slug with random suffix
function generateUniqueSlug(displayName: string): string {
  const baseSlug = generateSlug(displayName);
  const randomSuffix = Math.floor(100 + Math.random() * 900);

  if (!baseSlug) {
    const randomWords = ['workspace', 'team', 'org', 'company', 'studio'];
    const randomWord = randomWords[Math.floor(Math.random() * randomWords.length)];
    return `${randomWord}-${randomSuffix}`;
  }

  return `${baseSlug}-${randomSuffix}`;
}

export default function OrgSelectorPage() {
  const router = useRouter();
  const [memberships, setMemberships] = useState<MembershipInfo[]>([]);
  const [loading, setLoading] = useState(true);
  const [creatingOrg, setCreatingOrg] = useState(false);
  const [searchTerm, setSearchTerm] = useState("");
  const [userInfo, setUserInfo] = useState<{ name: string; phone: string } | null>(null);
  const { setCurrentOrg, setOrganizations, loadMemberships: loadMembershipsStore } = useOrganizationStore();

  useEffect(() => {
    const createDefaultOrganization = async (token: string, userData: { name: string; phone: string }) => {
      try {
        setCreatingOrg(true);

        const userName = userData.name || userData.phone;
        const displayName = `${userName}的工作空间`;

        console.log("[OrgSelector] Creating default organization:", { displayName });

        const response = await fetch("/api/org/init", {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
            Authorization: `Bearer ${token}`,
          },
          body: JSON.stringify({
            displayName: displayName,
          }),
        });

        const data = await response.json() as Record<string, string>;

        if (!response.ok) {
          console.error("[OrgSelector] Failed to initialize default org:", data.error);
          setCreatingOrg(false);
          setLoading(false);
          return;
        }

        console.log("[OrgSelector] Default organization initialized successfully:", data);

        const updatedMemberships = await loadMembershipsStore(token, true);
        console.log("[OrgSelector] Re-fetched memberships after creation:", updatedMemberships);

        setCreatingOrg(false);

        if (updatedMemberships.length > 0) {
          if (updatedMemberships.length === 1) {
            const org = updatedMemberships[0];
            setCurrentOrg(org.orgName);
            router.push(`/org/${org.orgName}/workspace`);
            return;
          }
          setMemberships(updatedMemberships);
        }
        setLoading(false);
      } catch (error) {
        console.error("[OrgSelector] Error creating default organization:", error);
        setCreatingOrg(false);
        setLoading(false);
      }
    };

    const loadMemberships = async () => {
      let token = getToken();
      console.log("[OrgSelector] In-memory token present:", !!token);

      // Restore access token if page was refreshed (middleware already checked cookie)
      if (!token) {
        console.log("[OrgSelector] Attempting silent refresh...");
        token = await refreshAccessToken();
        console.log("[OrgSelector] Silent refresh:", token ? "success" : "failed");
      }

      if (!token) {
        // Should not happen — middleware guards this
        console.warn("[OrgSelector] No token, redirecting to login");
        router.push("/login");
        return;
      }

      const userData = getUserInfoFromToken(token);

      if (!userData) {
        console.log("[OrgSelector] Failed to parse token, redirecting to login");
        removeToken();
        router.push("/login");
        return;
      }

      setUserInfo({
        name: userData.name,
        phone: userData.phone,
      });

      try {
        const apiMemberships = await loadMembershipsStore(token, true);

        console.log("[OrgSelector] Loaded memberships:", apiMemberships);

        if (apiMemberships.length === 0) {
          console.log("[OrgSelector] No organizations found (confirmed from backend), creating default org...");
          await createDefaultOrganization(token, userData);
          return;
        }

        const lastOrgId = localStorage.getItem('lastSelectedOrgId');
        const smartUrl = getSmartRedirectUrl(apiMemberships, lastOrgId || undefined);

        if (smartUrl !== '/org-selector') {
          console.log('[OrgSelector] Smart redirect to:', smartUrl);
          router.push(smartUrl);
          return;
        }

        setMemberships(apiMemberships);
        setLoading(false);
      } catch (error) {
        console.error("[OrgSelector] Error fetching memberships:", error);
        setMemberships([]);
        setLoading(false);
      }
    };

    loadMemberships();
  }, [router, loadMembershipsStore, setCurrentOrg]);

  const handleSelectOrg = (org: MembershipInfo) => {
    console.log("[OrgSelector] Selected org:", org.orgName);
    localStorage.setItem("lastSelectedOrgId", org.orgId);
    setCurrentOrg(org.orgName);
    router.push(`/org/${org.orgName}/workspace`);
  };

  const handleCreateOrg = () => {
    router.push("/org/create");
  };

  const handleSignOut = () => {
    removeToken();
    router.push("/login");
  };

  const handleRefresh = () => {
    setLoading(true);
    window.location.reload();
  };

  const filteredMemberships = memberships.filter((org) =>
    org.displayName.toLowerCase().includes(searchTerm.toLowerCase()) ||
    org.orgName.toLowerCase().includes(searchTerm.toLowerCase())
  );

  const displayName = userInfo?.name || userInfo?.phone || 'User';
  const initial = displayName.charAt(0).toUpperCase();

  const isOwnerOrAdmin = (role: string) =>
    role.toLowerCase() === 'owner' || role.toLowerCase() === 'admin';

  // ===== Loading State =====
  if (loading) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-[#fafafa]">
        <div className="text-center">
          <div className="mx-auto mb-3 size-9 animate-spin rounded-full border-2 border-gray-200 border-t-[#2563eb]" />
          <p className="text-sm text-[#6b7280]">
            {creatingOrg ? '正在为您创建工作空间...' : '加载组织列表...'}
          </p>
          {creatingOrg && (
            <p className="mt-1.5 text-xs text-[#9ca3af]">
              首次登录，系统正在自动创建您的默认工作空间
            </p>
          )}
        </div>
      </div>
    );
  }

  // ===== Main Page =====
  return (
    <div className="min-h-screen bg-[#fafafa]">

      {/* ===== Topbar ===== */}
      <header className="flex h-14 items-center justify-between border-b bg-white px-6">
        {/* Brand */}
        <div className="flex items-center gap-2.5">
          <div className="flex size-8 items-center justify-center rounded-lg bg-[#2563eb]">
            <Boxes className="size-4 text-white" strokeWidth={1.5} />
          </div>
          <span className="text-base font-semibold text-[#111827]">ModelCraft</span>
        </div>

        {/* User Menu */}
        {userInfo && (
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <button className="flex cursor-pointer items-center gap-2 rounded-md border-0 bg-transparent px-2 py-1.5 transition-colors hover:bg-[#fafafa]">
                <div className="flex size-7 items-center justify-center rounded-full bg-[#2563eb] text-xs font-semibold text-white">
                  {initial}
                </div>
                <span className="max-w-[120px] truncate text-sm font-medium text-[#111827]">
                  {displayName}
                </span>
                <svg className="size-3.5 text-[#9ca3af]" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
                  <path strokeLinecap="round" strokeLinejoin="round" d="M19 9l-7 7-7-7" />
                </svg>
              </button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end" className="w-52">
              <DropdownMenuLabel>
                <div>
                  <div className="text-sm font-semibold text-[#111827]">{userInfo.name}</div>
                  <div className="text-xs font-normal text-[#6b7280]">{userInfo.phone}</div>
                </div>
              </DropdownMenuLabel>
              <DropdownMenuSeparator />
              <DropdownMenuItem
                onClick={handleSignOut}
                className="cursor-pointer text-[#ef4444] focus:text-[#ef4444]"
              >
                <LogOut className="mr-2 size-3.5" strokeWidth={1.5} />
                退出登录
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        )}
      </header>

      {/* ===== Page Content ===== */}
      <div className="mx-auto max-w-4xl px-6 py-8">

        {/* Page Header Card */}
        <div className="mb-4 rounded-lg border bg-white p-5">
          <div className="mb-4 flex items-center justify-between">
            <div>
              <h1 className="flex items-center gap-2 text-lg font-semibold text-[#111827]">
                <Building2 className="size-[18px] text-[#2563eb]" strokeWidth={1.5} />
                选择组织
              </h1>
              <p className="mt-1 text-[13px] text-[#6b7280]">
                {memberships.length === 0
                  ? "创建您的第一个组织以开始使用 ModelCraft"
                  : `您有权访问 ${memberships.length} 个组织，请选择一个继续`}
              </p>
            </div>
            <div className="flex items-center gap-2">
              <button
                onClick={handleRefresh}
                disabled={loading}
                className="flex size-9 items-center justify-center rounded-md border bg-[#fafafa] text-[#6b7280] transition-all hover:border-[#d1d5db] hover:bg-white hover:text-[#111827] disabled:opacity-50"
                title="刷新"
              >
                <RefreshCw className={`size-4 ${loading ? 'animate-spin' : ''}`} strokeWidth={1.5} />
              </button>
              {memberships.length > 0 && (
                <Button
                  onClick={handleCreateOrg}
                  className="h-9 gap-1.5 border bg-[#fafafa] px-3 text-sm font-medium text-[#111827] shadow-none hover:border-[#d1d5db] hover:bg-white"
                  variant="outline"
                >
                  <Plus className="size-4" strokeWidth={1.5} />
                  创建组织
                </Button>
              )}
            </div>
          </div>

          {/* Search */}
          {memberships.length > 0 && (
            <div className="relative">
              <Search className="absolute left-2.5 top-1/2 size-3.5 -translate-y-1/2 text-[#9ca3af]" strokeWidth={1.5} />
              <Input
                placeholder="搜索组织名称..."
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                className="h-9 bg-[#fafafa] pl-8 text-sm placeholder:text-[#9ca3af] focus:bg-white"
              />
            </div>
          )}
        </div>

        {/* ===== Org Grid ===== */}
        {memberships.length > 0 ? (
          filteredMemberships.length > 0 ? (
            <div className="grid grid-cols-1 gap-3 md:grid-cols-2 lg:grid-cols-3">
              {filteredMemberships.map((org) => (
                <div
                  key={org.orgId}
                  role="button"
                  tabIndex={0}
                  onClick={() => handleSelectOrg(org)}
                  onKeyDown={(e) => {
                    if (e.key === "Enter" || e.key === " ") {
                      e.preventDefault();
                      handleSelectOrg(org);
                    }
                  }}
                  className="group cursor-pointer rounded-lg border bg-white p-4 transition-all hover:border-[#bfdbfe] hover:shadow-sm"
                >
                  {/* Card Header */}
                  <div className="mb-3 flex items-start gap-3">
                    <div className="flex size-9 flex-shrink-0 items-center justify-center rounded-lg bg-[#2563eb]">
                      <Building2 className="size-4 text-white" strokeWidth={1.5} />
                    </div>
                    <div className="min-w-0 flex-1">
                      <div className="truncate text-[15px] font-semibold text-[#111827]">
                        {org.displayName}
                      </div>
                      <div className="font-mono text-[12px] text-[#9ca3af]">
                        @{org.orgName}
                      </div>
                    </div>
                    {/* Enter Arrow */}
                    <div className="flex size-7 flex-shrink-0 items-center justify-center rounded-md border bg-[#fafafa] transition-all group-hover:border-[#2563eb] group-hover:bg-[#2563eb]">
                      <ArrowRight
                        className="size-3.5 text-[#6b7280] transition-colors group-hover:text-white"
                        strokeWidth={1.5}
                      />
                    </div>
                  </div>

                  {/* Card Footer */}
                  <div className="flex items-center justify-between border pt-3">
                    {isOwnerOrAdmin(org.role) ? (
                      <span className="inline-flex items-center gap-1 rounded bg-[#dbeafe] px-2.5 py-1 text-xs font-medium text-[#2563eb]">
                        <ShieldCheck className="size-3" strokeWidth={1.5} />
                        {org.role.charAt(0).toUpperCase() + org.role.slice(1).toLowerCase()}
                      </span>
                    ) : (
                      <span className="inline-flex items-center gap-1 rounded bg-[#f3f4f6] px-2.5 py-1 text-xs font-medium text-[#6b7280]">
                        <Users className="size-3" strokeWidth={1.5} />
                        {org.role.charAt(0).toUpperCase() + org.role.slice(1).toLowerCase()}
                      </span>
                    )}
                    <span className="text-[12px] text-[#9ca3af]">
                      {new Date(org.joinedAt).toLocaleDateString('zh-CN', {
                        year: 'numeric',
                        month: 'short',
                        day: 'numeric',
                      })}
                    </span>
                  </div>
                </div>
              ))}
            </div>
          ) : (
            /* No search results */
            <div className="rounded-lg border bg-white px-6 py-16 text-center">
              <div className="mx-auto mb-4 flex size-12 items-center justify-center rounded-xl border bg-[#fafafa]">
                <Building2 className="size-5 text-[#9ca3af]" strokeWidth={1.5} />
              </div>
              <p className="text-[15px] font-semibold text-[#111827]">未找到匹配的组织</p>
              <p className="mt-1.5 text-[13px] text-[#6b7280]">尝试使用其他关键词搜索</p>
            </div>
          )
        ) : (
          /* Empty state */
          <div className="rounded-lg border bg-white px-6 py-16 text-center">
            <div className="mx-auto mb-4 flex size-12 items-center justify-center rounded-xl border bg-[#fafafa]">
              <Building2 className="size-5 text-[#9ca3af]" strokeWidth={1.5} />
            </div>
            <p className="text-[15px] font-semibold text-[#111827]">暂无组织</p>
            <p className="mx-auto mt-1.5 max-w-sm text-[13px] text-[#6b7280]">
              创建您的第一个组织以开始使用 ModelCraft 构建数据驱动的应用
            </p>
            <Button
              onClick={handleCreateOrg}
              className="mt-5 h-9 gap-1.5 bg-[#2563eb] px-4 text-sm font-medium text-white hover:bg-[#1d4ed8]"
            >
              <Plus className="size-4" strokeWidth={1.5} />
              创建组织
            </Button>
          </div>
        )}
      </div>
    </div>
  );
}
