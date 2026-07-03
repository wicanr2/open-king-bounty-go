package com.wicanr2.openkb;

import android.app.Activity;
import android.os.Bundle;
import android.view.View;
import android.view.Window;
import android.view.WindowManager;

import com.wicanr2.openkb.mobile.EbitenView;

public class MainActivity extends Activity {
    private EbitenView getEbitenView() {
        return (EbitenView) this.findViewById(R.id.ebitenview);
    }

    // enterImmersive 隱藏狀態列 + 導覽列,讓遊戲填滿整個螢幕(對齊 C openkb 的全螢幕
    // 呈現;也讓觸控座標換算不必扣掉系統列高度)。IMMERSIVE_STICKY:玩家從邊緣滑出
    // 系統列後會自動再隱藏。需在 onCreate 與每次取得焦點(onWindowFocusChanged)時重設。
    private void enterImmersive() {
        View decor = getWindow().getDecorView();
        decor.setSystemUiVisibility(
            View.SYSTEM_UI_FLAG_IMMERSIVE_STICKY
            | View.SYSTEM_UI_FLAG_FULLSCREEN
            | View.SYSTEM_UI_FLAG_HIDE_NAVIGATION
            | View.SYSTEM_UI_FLAG_LAYOUT_STABLE
            | View.SYSTEM_UI_FLAG_LAYOUT_FULLSCREEN
            | View.SYSTEM_UI_FLAG_LAYOUT_HIDE_NAVIGATION);
    }

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        // 無標題列 + 全螢幕旗標(須在 setContentView 前設)。
        requestWindowFeature(Window.FEATURE_NO_TITLE);
        getWindow().setFlags(WindowManager.LayoutParams.FLAG_FULLSCREEN,
            WindowManager.LayoutParams.FLAG_FULLSCREEN);
        // 把 Application context 交給 gomobile —— ebitenmobile bind 是嵌在一般 Activity
        // 的 view,沒有 NativeActivity 流程幫忙設 context;不設的話 gomobile 的 RunOnJVM
        // 拿不到 context,ebiten 的 devicescale 查詢失敗回 0 → view 尺寸變 +Inf → 全黑。
        go.Seq.setContext(getApplicationContext());
        setContentView(R.layout.activity_main);
        enterImmersive();
    }

    @Override
    public void onWindowFocusChanged(boolean hasFocus) {
        super.onWindowFocusChanged(hasFocus);
        if (hasFocus) {
            enterImmersive();
        }
    }

    @Override
    protected void onPause() {
        super.onPause();
        getEbitenView().suspendGame();
    }

    @Override
    protected void onResume() {
        super.onResume();
        getEbitenView().resumeGame();
    }
}
