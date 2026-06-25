import { IconSettings } from "@tabler/icons-react";
import { useState } from "react";

import { AppConfig } from "./_components/app-config";

const settingTabs = [
  {
    key: "app",
    title: "应用设置",
    icon: IconSettings,
    description: "应用行为偏好配置",
  },
];

export function AppSettingPage() {
  const [activeTab, setActiveTab] = useState("app");

  return (
    <div className="h-screen p-6">
      {/* 页面标题 */}
      <div className="mb-6">
        <h1 className="text-2xl font-bold mb-2">系统设置</h1>
        <p className="text-gray-600 dark:text-gray-400">配置系统偏好设置</p>
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
          <div className="space-y-6">{activeTab === "app" && <AppConfig />}</div>
        </div>
      </div>
    </div>
  );
}
