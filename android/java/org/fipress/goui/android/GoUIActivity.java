package org.fipress.goui.android;

import android.provider.ContactsContract;
import android.app.Activity;
import android.os.Bundle;
import android.util.Log;
import android.webkit.ValueCallback;
import android.webkit.WebSettings;
import android.webkit.WebView;

import java.time.LocalDateTime;
import java.util.Date;

public class GoUIActivity extends Activity {

    // Used to load the 'native-lib' library on application startup.
    static {
        System.loadLibrary("goui");
    }

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        getActionBar().hide();

        webView = new WebView(this);
        WebSettings webSettings = webView.getSettings();
        webSettings.setJavaScriptEnabled(true);
        webView.addJavascriptInterface(new ScriptHandler(this),"gouiAndroid");
        setContentView(webView);

        if(webUrl == null || webUrl.isEmpty()) {
            webView.loadUrl(defaultUrl);
            if(!invokeGoMain()) {
                throw new RuntimeException("Invoke main function failed.");
            }
        } else {
            webView.loadUrl(webUrl);
        }
        //String url = setUrl();
        //System.out.println("set url:"+url);
        //boolean debug = setDebug();
        //System.out.println(debug);
    }

    private void loadWebView(String url) {
        Log.d(logTag,"create web view, url:"+url);
        webUrl = url;
        if(!webUrl.toLowerCase().equals(defaultUrl)) {
            webView.loadUrl(url);
        }

        //if(debugEnabled) { //todo
        WebView.setWebContentsDebuggingEnabled(true);
        //}
    }

    public void evalJavaScript(final String script) {
        webView.post(new Runnable() {
            @Override
            public void run() {
                webView.evaluateJavascript(script, new ValueCallback<String>() {
                    @Override
                    public void onReceiveValue(String value) {
                        Log.i(logTag,value);
                    }
                });
            }
        });
    }

    //invoke main.main from Go
    public native boolean invokeGoMain();
}
