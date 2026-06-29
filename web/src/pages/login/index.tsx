import augmentedRealityImage from "@/assets/login-img/undraw_augmented-reality_3ie0.svg";
import outerSpaceImage from "@/assets/login-img/undraw_outer-space_qey5.svg";
import pictureImage from "@/assets/login-img/undraw_picture_lnem.svg";
import starsImage from "@/assets/login-img/undraw_stars_5pgw.svg";
import toTheMoonImage from "@/assets/login-img/undraw_to-the-moon_w1wa.svg";
import logo from "@/assets/logo.svg";

import { LoginForm } from "./login-form";

export function LoginPage() {
  return (
    <main className="relative min-h-screen overflow-hidden bg-app-shell">
      <img
        src={starsImage}
        alt=""
        aria-hidden="true"
        className="absolute left-1/2 top-1/2 h-auto w-[min(92vw,1040px)] -translate-x-1/2 -translate-y-1/2 opacity-95"
      />
      <div className="absolute inset-0 bg-[radial-gradient(circle_at_50%_45%,rgb(255_255_255/0.92),rgb(238_247_255/0.72)_38%,rgb(238_247_255/0.3)_58%,transparent_78%)] dark:bg-[radial-gradient(circle_at_50%_45%,rgb(17_24_39/0.72),rgb(17_24_39/0.52)_40%,rgb(17_24_39/0.18)_64%,transparent_82%)]" />
      <img
        src={toTheMoonImage}
        alt=""
        aria-hidden="true"
        className="absolute -bottom-10 left-1/2 hidden w-[260px] -translate-x-[243%] -translate-y-[40%] opacity-90 drop-shadow-[0_22px_56px_rgb(15_23_42/0.18)] md:block lg:w-[330px] xl:w-[390px]"
      />
      <img
        src={outerSpaceImage}
        alt=""
        aria-hidden="true"
        className="absolute right-4 top-8 hidden w-[210px] rotate-6 rounded-lg bg-white/78 p-3 opacity-92 shadow-[0_22px_56px_rgb(15_23_42/0.14)] backdrop-blur-sm sm:block lg:right-24 lg:top-14 lg:w-[270px] dark:bg-white/10"
      />
      <img
        src={augmentedRealityImage}
        alt=""
        aria-hidden="true"
        className="absolute left-3 top-24 hidden w-[230px] -rotate-3 rounded-lg bg-white/78 p-3 opacity-92 shadow-[0_22px_56px_rgb(15_23_42/0.14)] backdrop-blur-sm md:block lg:left-24 lg:top-20 lg:w-[300px] dark:bg-white/10"
      />
      <img
        src={pictureImage}
        alt=""
        aria-hidden="true"
        className="absolute bottom-10 right-8 hidden w-[190px] rotate-2 rounded-lg bg-white/80 p-3 opacity-92 shadow-[0_20px_52px_rgb(15_23_42/0.14)] backdrop-blur-sm md:block lg:right-28 lg:w-[240px] dark:bg-white/10"
      />
      <div className="absolute inset-x-0 bottom-0 h-1/2 bg-gradient-to-t from-background/82 via-background/34 to-transparent dark:from-background/90" />

      <div className="relative z-10 flex min-h-screen w-full items-center justify-center px-6 py-10 sm:px-8">
        <section className="w-full max-w-sm">
          <div className="mb-6 flex items-center gap-3">
            <img src={logo} alt="Verve" className="size-10 shrink-0 rounded-lg object-contain" />
            <div>
              <h1 className="text-foreground text-xl font-bold">Verve</h1>
              <p className="text-muted-foreground text-xs">你的 AI 私教，陪你真正学会</p>
            </div>
          </div>

          <LoginForm />
        </section>
      </div>
    </main>
  );
}
