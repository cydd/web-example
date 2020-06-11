import { BrowserModule } from '@angular/platform-browser';
import { NgModule } from '@angular/core';

import { AppRoutingModule } from './app-routing.module';
import { AppComponent } from './app.component';
import { NgZorroAntdModule, NZ_I18N, zh_CN } from 'ng-zorro-antd';
import { FormsModule, ReactiveFormsModule } from '@angular/forms';
import { HttpClientModule } from '@angular/common/http';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { registerLocaleData } from '@angular/common';
import { JwtModule } from '@auth0/angular-jwt';
import zh from '@angular/common/locales/zh';
import { RecaptchaModule } from 'ng-recaptcha';

import { HeaderComponent } from './layout/header/header.component';
import { FooterComponent } from './layout/footer/footer.component';
import { AuthComponent } from './pages/auth/auth.component';
import { RegisterComponent } from './pages/register/register.component';
import { UserListComponent } from './pages/userMgr/userlist/userlist.component';
import { adminDashboardComponent } from './pages/userMgr/dashboard/dashboard.component';
import { AdduserComponent } from './pages/userMgr/adduser/adduser.component';
import { DashboardComponent } from './pages/dashboard/dashboard.component';
import { Guard } from './guard';
import { MessageComponent } from './pages/message/message.component';
import { NbThemeModule, NbLayoutModule, NbChatModule } from '@nebular/theme';
import { NbEvaIconsModule } from '@nebular/eva-icons'

registerLocaleData(zh);

export function tokenGetter() {
  return localStorage.getItem('token');
}

@NgModule({
  declarations: [
    AppComponent,
    HeaderComponent,
    FooterComponent,
    AuthComponent,
    RegisterComponent,
    UserListComponent,
    adminDashboardComponent,
    AdduserComponent,
    DashboardComponent,
    MessageComponent,
  ],
  imports: [
    NbChatModule,
    RecaptchaModule,
    BrowserModule,
    AppRoutingModule,
    NgZorroAntdModule,
    FormsModule,
    HttpClientModule,
    BrowserAnimationsModule,
    ReactiveFormsModule,
    JwtModule.forRoot({
      config: {
        tokenGetter: tokenGetter
      }
    }),
    NbThemeModule.forRoot({ name: 'default' }),
    NbLayoutModule,
    NbEvaIconsModule
  ],
  providers: [ Guard, { provide: NZ_I18N, useValue: zh_CN } ],
  bootstrap: [AppComponent]
})
export class AppModule { }
