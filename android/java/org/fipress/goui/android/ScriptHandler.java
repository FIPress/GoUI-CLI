package org.fipress.goui.android;

import android.content.Context;
import android.util.Log;
import android.webkit.JavascriptInterface;

public class ScriptHandler {
    private Context context;

    public ScriptHandler(Context _context) {
        context = _context;
    }

    @JavascriptInterface
    public void handleMessage(String message) {
        System.out.println(message);
        Log.d(GoUIActivity.logTag,"script handler:",message);
        postMessage(message);
    }

    public native String postMessage(String message);

}
