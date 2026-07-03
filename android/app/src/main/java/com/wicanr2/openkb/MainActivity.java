package com.wicanr2.openkb;

import android.app.Activity;
import android.os.Bundle;

import com.wicanr2.openkb.mobile.EbitenView;

public class MainActivity extends Activity {
    private EbitenView getEbitenView() {
        return (EbitenView) this.findViewById(R.id.ebitenview);
    }

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        // 把 Application context 交給 gomobile —— ebitenmobile bind 是嵌在一般 Activity
        // 的 view,沒有 NativeActivity 流程幫忙設 context;不設的話 gomobile 的 RunOnJVM
        // 拿不到 context,ebiten 的 devicescale 查詢失敗回 0 → view 尺寸變 +Inf → 全黑。
        go.Seq.setContext(getApplicationContext());
        setContentView(R.layout.activity_main);
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
