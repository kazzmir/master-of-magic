package org.masterofmagic.arena;

import android.app.Activity;
import android.app.AlertDialog;
import android.os.Bundle;
import android.view.View;
import go.Seq;
import org.masterofmagic.arena.mobile.EbitenView;

public class MainActivity extends Activity {
    private EbitenView gameView;

    @Override
    protected void onCreate(Bundle state) {
        super.onCreate(state);
        Seq.setContext(getApplicationContext());
        gameView = new EbitenView(this) {
            @Override
            protected void onErrorOnGameUpdate(Exception error) {
                new AlertDialog.Builder(MainActivity.this)
                    .setTitle("Arena error")
                    .setMessage(error.toString())
                    .setPositiveButton("Close", (dialog, which) -> finish())
                    .show();
            }
        };
        setContentView(gameView);
        getWindow().getDecorView().setSystemUiVisibility(
            View.SYSTEM_UI_FLAG_FULLSCREEN | View.SYSTEM_UI_FLAG_HIDE_NAVIGATION
            | View.SYSTEM_UI_FLAG_IMMERSIVE_STICKY);
    }

    @Override
    protected void onPause() {
        gameView.suspendGame();
        super.onPause();
    }

    @Override
    protected void onResume() {
        super.onResume();
        if (gameView != null) {
            gameView.resumeGame();
        }
    }
}
