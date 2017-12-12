package com.swisscom.bitbucket;

import com.intuit.karate.junit4.Karate;
import org.junit.AfterClass;
import org.junit.BeforeClass;
import org.junit.runner.RunWith;

@RunWith(Karate.class)
public abstract class TestBase {
    
    @BeforeClass
    public static void beforeClass() throws Exception {
    }
    
    @AfterClass
    public static void afterClass() {
    }
    
}
