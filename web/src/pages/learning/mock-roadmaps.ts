import type { Edge, Node } from "@xyflow/react";

export type RoadmapStageStatus = "planned" | "active" | "completed";

export type RoadmapLesson = {
  id: string;
  title: string;
  summary: string;
  outcomes: string[];
  duration: string;
  resources: string[];
  tasks: string[];
};

export type RoadmapStage = {
  id: string;
  title: string;
  summary: string;
  description: string;
  duration: string;
  difficulty: "入门" | "进阶" | "强化";
  status: RoadmapStageStatus;
  order: number;
  x: number;
  y: number;
  outcomes: string[];
  lessons: RoadmapLesson[];
};

export type LearningRoadmap = {
  id: string;
  title: string;
  description: string;
  tagline: string;
  level: string;
  duration: string;
  progress: number;
  learners: string;
  category: "frontend" | "engineering" | "ai";
  tags: string[];
  heroPoints: string[];
  stages: RoadmapStage[];
};

type RoadmapNodeData = {
  label: string;
  duration: string;
  difficulty: string;
  status: RoadmapStageStatus;
};

export const learningRoadmaps: LearningRoadmap[] = [
  {
    id: "frontend-system",
    title: "前端工程化进阶",
    description: "从组件抽象、状态管理到构建发布，搭起一条能真正落地的现代前端工程路线。",
    tagline: "适合已经会 React，但项目一大就开始乱的人。",
    level: "React 基础后",
    duration: "6 周",
    progress: 42,
    learners: "126 人在学",
    category: "engineering",
    tags: ["React", "TypeScript", "工程化", "组件设计"],
    heroPoints: ["做出可复用组件体系", "把状态和数据流理顺", "建立构建发布认知"],
    stages: [
      {
        id: "fe-foundation",
        title: "组件与状态基础",
        summary: "先把页面拆分能力和状态边界感建立起来。",
        description: "用真实页面拆分练习，把 UI、业务状态和数据状态分开，避免页面一改就全连着动。",
        duration: "第 1-2 周",
        difficulty: "入门",
        status: "completed",
        order: 1,
        x: 60,
        y: 80,
        outcomes: ["会做组件职责划分", "知道局部状态与全局状态怎么选"],
        lessons: [
          {
            id: "fe-foundation-1",
            title: "组件拆分的边界感",
            summary: "把一个复杂页面拆成结构组件、功能组件和可复用组件。",
            outcomes: ["识别 props 膨胀", "识别可复用片段"],
            duration: "2 小时",
            resources: ["Dashboard 页面拆分案例", "组件设计 checklist"],
            tasks: ["拆一个列表页", "画出组件层级图"],
          },
          {
            id: "fe-foundation-2",
            title: "状态放在哪最合适",
            summary: "区分局部 UI 状态、共享状态和服务端状态。",
            outcomes: ["知道状态提升何时发生", "避免无意义全局 store"],
            duration: "2.5 小时",
            resources: ["表单状态案例", "查询缓存示例"],
            tasks: ["重构一个弹窗流程", "给状态分类"],
          },
        ],
      },
      {
        id: "fe-dataflow",
        title: "数据流与交互组织",
        summary: "把页面联动和请求流程组织得清清楚楚。",
        description: "重点练查询、变更、乐观更新、错误反馈这些真实项目最常见的交互链路。",
        duration: "第 3-4 周",
        difficulty: "进阶",
        status: "active",
        order: 2,
        x: 340,
        y: 40,
        outcomes: ["能设计稳定的数据流", "能处理复杂交互反馈"],
        lessons: [
          {
            id: "fe-dataflow-1",
            title: "查询与变更的页面编排",
            summary: "让 loading、empty、error、success 各归其位。",
            outcomes: ["能设计页面状态矩阵", "减少 if else 泥团"],
            duration: "3 小时",
            resources: ["Query 模式卡", "状态矩阵模板"],
            tasks: ["补齐一个页面的状态图", "整理交互优先级"],
          },
          {
            id: "fe-dataflow-2",
            title: "复杂交互的节奏控制",
            summary: "处理草稿、自动保存、撤销和操作确认。",
            outcomes: ["懂得交互反馈节奏", "会做最小可理解操作流"],
            duration: "2 小时",
            resources: ["操作流示例", "反馈组件使用方式"],
            tasks: ["重写一个表单提交流程", "加上撤销反馈"],
          },
        ],
      },
      {
        id: "fe-delivery",
        title: "构建质量与交付",
        summary: "从开发者视角走到交付视角。",
        description: "补上 lint、打包、环境、性能和发布这条链，让项目不是只能在本地跑。",
        duration: "第 5-6 周",
        difficulty: "强化",
        status: "planned",
        order: 3,
        x: 640,
        y: 120,
        outcomes: ["能读懂构建链", "知道上线前重点查什么"],
        lessons: [
          {
            id: "fe-delivery-1",
            title: "开发到发布的路径图",
            summary: "理解从本地构建到线上发布的关键节点。",
            outcomes: ["会定位构建链问题", "理解环境变量与构建产物"],
            duration: "2 小时",
            resources: ["CI/CD 流程图", "打包分析示例"],
            tasks: ["画一张交付流程图", "记录发布风险点"],
          },
          {
            id: "fe-delivery-2",
            title: "性能与质量守门",
            summary: "知道哪些问题适合靠 lint、哪些适合靠代码评审或运行验证。",
            outcomes: ["建立质量分层意识", "能设计最小验证闭环"],
            duration: "2.5 小时",
            resources: ["性能排查清单", "发布前检查项"],
            tasks: ["做一次性能走查", "写一版上线检查单"],
          },
        ],
      },
    ],
  },
  {
    id: "react-product",
    title: "React 产品实战路线",
    description: "围绕真实产品页面，从信息架构、交互拆解到功能落地，一步步做出完整前端作品。",
    tagline: "适合想摆脱 demo 感、开始做像样产品的人。",
    level: "有 React 基础",
    duration: "4 周",
    progress: 18,
    learners: "89 人在学",
    category: "frontend",
    tags: ["React", "产品思维", "交互设计", "页面组织"],
    heroPoints: ["会做完整页面流", "理解信息层次", "知道如何把需求落成页面"],
    stages: [
      {
        id: "rp-brief",
        title: "需求拆解与页面结构",
        summary: "先把产品目标翻译成页面结构。",
        description: "学会从一句需求里抽出主任务、次任务、页面优先级，再去落布局。",
        duration: "第 1 周",
        difficulty: "入门",
        status: "active",
        order: 1,
        x: 80,
        y: 90,
        outcomes: ["会写页面结构草图", "会拆主次任务"],
        lessons: [
          {
            id: "rp-brief-1",
            title: "一句需求怎么拆页面",
            summary: "从目标、角色、任务和动作四个角度看需求。",
            outcomes: ["能写页面骨架", "不再一上来就写组件"],
            duration: "90 分钟",
            resources: ["需求拆解模板", "页面草图案例"],
            tasks: ["拆一个学习页", "写出 3 层信息结构"],
          },
        ],
      },
      {
        id: "rp-build",
        title: "关键页面与状态联动",
        summary: "做出可点击、可操作、可迭代的主要页面。",
        description: "聚焦列表页、详情页、操作面板之间的节奏和状态同步。",
        duration: "第 2-3 周",
        difficulty: "进阶",
        status: "planned",
        order: 2,
        x: 380,
        y: 30,
        outcomes: ["理解页面间流转", "知道如何组织操作面板"],
        lessons: [
          {
            id: "rp-build-1",
            title: "列表页到详情页的叙事感",
            summary: "让页面切换不是跳转，而是信息延续。",
            outcomes: ["会做入口卡片", "会在详情页延续上下文"],
            duration: "2 小时",
            resources: ["列表详情结构样例", "交互流示意"],
            tasks: ["重做一个卡片入口", "优化详情首屏"],
          },
        ],
      },
      {
        id: "rp-polish",
        title: "复盘与作品化",
        summary: "把做出来的东西整理成能展示的成品。",
        description: "整理设计取舍、结构图和关键交互，让项目不只是代码，而是完整作品。",
        duration: "第 4 周",
        difficulty: "强化",
        status: "planned",
        order: 3,
        x: 680,
        y: 120,
        outcomes: ["知道如何讲清楚项目", "会输出项目说明"],
        lessons: [
          {
            id: "rp-polish-1",
            title: "把项目讲明白",
            summary: "围绕问题、方案、结果组织你的项目叙述。",
            outcomes: ["会写项目说明", "会做结果导向复盘"],
            duration: "90 分钟",
            resources: ["作品集说明模板", "复盘问题清单"],
            tasks: ["写一版项目复盘", "整理关键截图与结构图"],
          },
        ],
      },
    ],
  },
  {
    id: "ai-builder",
    title: "AI 应用前端路线图",
    description: "面向 AI 产品的前端能力：对话流、可视化工作流、多状态反馈和工具调用界面。",
    tagline: "适合准备做 AI 工作台、Agent 页面或多轮交互产品的人。",
    level: "中级前端",
    duration: "5 周",
    progress: 0,
    learners: "57 人在学",
    category: "ai",
    tags: ["AI 产品", "多轮交互", "工作流", "可视化"],
    heroPoints: ["理解 AI 产品前端模式", "能做多状态会话界面", "能设计工作流视图"],
    stages: [
      {
        id: "ai-chat",
        title: "对话式交互基础",
        summary: "先理解 AI 界面的基本状态和输入输出节奏。",
        description: "围绕消息、输入、生成中状态、工具调用提示这些基础反馈建立认知。",
        duration: "第 1-2 周",
        difficulty: "入门",
        status: "planned",
        order: 1,
        x: 70,
        y: 90,
        outcomes: ["理解生成式 UI 反馈层次", "能做会话页基本布局"],
        lessons: [
          {
            id: "ai-chat-1",
            title: "会话界面的状态设计",
            summary: "设计等待、生成、失败、重试和引用信息。",
            outcomes: ["会做状态反馈", "理解消息结构"],
            duration: "2 小时",
            resources: ["聊天 UI 示例", "消息状态 checklist"],
            tasks: ["设计消息卡片状态", "梳理输入区功能"],
          },
        ],
      },
      {
        id: "ai-flow",
        title: "可视化流程与节点表达",
        summary: "进入你这次最想要的路线图/思维导图表达方式。",
        description: "学习如何把抽象步骤、节点依赖和任务流转展示成可交互的可视化画布。",
        duration: "第 3 周",
        difficulty: "进阶",
        status: "planned",
        order: 2,
        x: 360,
        y: 20,
        outcomes: ["理解节点编排", "会做可视化画布入口"],
        lessons: [
          {
            id: "ai-flow-1",
            title: "节点与连线的叙事方式",
            summary: "不是为了炫，而是为了让用户一眼看懂下一步。",
            outcomes: ["能设计节点层级", "能规划画布焦点"],
            duration: "2 小时",
            resources: ["流程图案例", "节点信息层级模板"],
            tasks: ["画一版学习路线节点", "决定每个节点展示字段"],
          },
        ],
      },
      {
        id: "ai-workbench",
        title: "AI 工作台整合",
        summary: "把会话、路线图、资源和任务板拼成一个整体。",
        description:
          "理解多面板布局、上下文同步和工具区组织方式，让 AI 产品前端更像工作台而不是单聊天框。",
        duration: "第 4-5 周",
        difficulty: "强化",
        status: "planned",
        order: 3,
        x: 660,
        y: 130,
        outcomes: ["能组织工作台界面", "理解多视图协作"],
        lessons: [
          {
            id: "ai-workbench-1",
            title: "从聊天框到工作台",
            summary: "让不同面板围绕一个目标协同工作。",
            outcomes: ["会做布局编排", "理解上下文共享"],
            duration: "3 小时",
            resources: ["工作台案例", "多面板布局参考"],
            tasks: ["画一版工作台布局", "梳理面板间共享状态"],
          },
        ],
      },
    ],
  },
];

export function getRoadmapList() {
  return learningRoadmaps;
}

export function getRoadmapById(id: string) {
  return learningRoadmaps.find((roadmap) => roadmap.id === id);
}

export function getRoadmapFlow(roadmap: LearningRoadmap): {
  nodes: Node<RoadmapNodeData>[];
  edges: Edge[];
} {
  const nodes = roadmap.stages.map((stage) => ({
    id: stage.id,
    type: "roadmapNode",
    position: { x: stage.x, y: stage.y },
    data: {
      label: stage.title,
      duration: stage.duration,
      difficulty: stage.difficulty,
      status: stage.status,
    },
  }));

  const edges = roadmap.stages.slice(1).map((stage, index) => ({
    id: `${roadmap.stages[index]?.id}-${stage.id}`,
    source: roadmap.stages[index]?.id ?? stage.id,
    target: stage.id,
    animated: stage.status === "active",
    style: {
      stroke: stage.status === "completed" ? "var(--primary)" : "var(--border)",
      strokeWidth: 2,
    },
  }));

  return { nodes, edges };
}
