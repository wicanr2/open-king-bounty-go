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
