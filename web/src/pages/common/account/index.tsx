import { IconLock, IconSettings } from "@tabler/icons-react";
import { useState } from "react";

import { AccountConfig } from "./_components/account-config";
import { PasswordConfig } from "./_components/password-config";

const settingTabs = [
  {
    key: "account",
    title: "账号设置",
    icon: IconSettings,
    description: "配置账号信息",
  },
  {
    key: "password",
    title: "密码设置",
    icon: IconLock,
    description: "配置密码信息",
  },
];

export function AccountPage() {
  const [activeTab, setActiveTab] = useState("account");

  return (
    <div className="h-screen p-6">
      {/* 页面标题 */}
      <div className="mb-6">
        <h1 className="text-2xl font-bold mb-2">账号设置</h1>
        <p className="text-gray-600 dark:text-gray-400">配置账号信息</p>
      </div>

      {/* 主要内容区域 */}
      <div className="flex gap-6">
        {/* 左侧导航 */}
        <div className="w-64 shrink-0">
          <div className="sticky top-6">
            <div className="space-y-2">
              {settingTabs.map((tab) => {
                const Icon = tab.icon;
                const isActive = activeTab === tab.key;
                return (
                  <div
                    key={tab.key}
                    className={`flex items-start gap-3 p-3 rounded-lg cursor-pointer transition-all duration-200 ${
                      isActive
                        ? "bg-primary text-primary-foreground shadow-sm"
                        : "hover:bg-accent hover:text-accent-foreground"
                    }`}
                    onClick={() => setActiveTab(tab.key)}
                  >
                    <Icon className="h-5 w-5 mt-0.5 shrink-0" />
                    <div className="flex-1 min-w-0">
                      <div className="font-medium text-sm">{tab.title}</div>
                      <div className="text-xs opacity-70 mt-1 line-clamp-2">{tab.description}</div>
                    </div>
                  </div>
                );
              })}
            </div>
          </div>
        </div>

        {/* 右侧配置内容 */}
        <div className="flex-1 min-w-0">
          <div className="space-y-6">
            {activeTab === "account" && <AccountConfig />}
            {activeTab === "password" && <PasswordConfig />}
          </div>
        </div>
      </div>
    </div>
  );
}
