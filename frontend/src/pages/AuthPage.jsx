import { useMemo, useState } from 'react';
import { toast } from 'react-hot-toast';

import { loginSchema, registerSchema } from '../utils/validators';

export function AuthPage({ classes, onLogin, onRegister, loading }) {
  const [activeTab, setActiveTab] = useState('login');
  const [loginForm, setLoginForm] = useState({ username: '', password: '' });
  const [registerForm, setRegisterForm] = useState({ username: '', password: '', classId: '' });

  const classOptions = useMemo(() => classes || [], [classes]);

  const submitLogin = async (event) => {
    event.preventDefault();
    try {
      const payload = loginSchema.parse(loginForm);
      await onLogin(payload);
    } catch (error) {
      toast.error(error?.issues?.[0]?.message || error.message || '登录失败');
    }
  };

  const submitRegister = async (event) => {
    event.preventDefault();
    try {
      const payload = registerSchema.parse({
        ...registerForm,
        classId: Number(registerForm.classId),
      });
      await onRegister(payload);
    } catch (error) {
      toast.error(error?.issues?.[0]?.message || error.message || '注册失败');
    }
  };

  return (
    <div className="min-h-screen bg-board px-4 py-8 md:py-14">
      <header className="mx-auto mb-8 max-w-5xl rounded-3xl border border-teal-100 bg-white/80 p-6 shadow-card backdrop-blur">
        <div className="flex flex-col gap-3 md:flex-row md:items-end md:justify-between">
          <div>
            <p className="text-sm font-semibold uppercase tracking-[0.25em] text-teal-700">Computer Lab Quiz Hub</p>
            <h1 className="mt-2 text-3xl font-bold text-slate-800">电脑室教师机答题系统</h1>
            <p className="mt-2 text-sm text-slate-600">教师端集中管理题库与班级错题统计，学生端注册后在线答题并自动生成错题本。</p>
          </div>
        </div>
      </header>

      <main className="mx-auto grid max-w-5xl gap-6 md:grid-cols-[1.2fr,1fr]">
        <section className="rounded-3xl border border-slate-200 bg-white/90 p-6 shadow-card backdrop-blur">
          <h2 className="text-xl font-semibold text-slate-800">系统能力</h2>
          <ul className="mt-4 space-y-3 text-sm text-slate-600">
            <li>题库支持新增、编辑、删除、JSON 上传。</li>
            <li>学生按班级注册登录，教师端可查看班级错题热区。</li>
            <li>题目选项每次下发都会随机排序，防止固定记忆。</li>
            <li>答题成绩实时写入数据库并同步到教师看板。</li>
            <li>学生自动沉淀错题本，支持针对性复盘。</li>
          </ul>
        </section>

        <section className="rounded-3xl border border-slate-200 bg-white p-6 shadow-card">
          <div className="tabs tabs-boxed mb-4 bg-slate-100 p-1">
            <button
              type="button"
              className={`tab flex-1 ${activeTab === 'login' ? 'tab-active' : ''}`}
              onClick={() => setActiveTab('login')}
            >
              登录
            </button>
            <button
              type="button"
              className={`tab flex-1 ${activeTab === 'register' ? 'tab-active' : ''}`}
              onClick={() => setActiveTab('register')}
            >
              学生注册
            </button>
          </div>

          {activeTab === 'login' ? (
            <form className="space-y-3" onSubmit={submitLogin}>
              <label className="form-control">
                <span className="label-text mb-1">用户名</span>
                <input
                  className="input input-bordered"
                  value={loginForm.username}
                  onChange={(event) => setLoginForm((prev) => ({ ...prev, username: event.target.value }))}
                />
              </label>
              <label className="form-control">
                <span className="label-text mb-1">密码</span>
                <input
                  type="password"
                  className="input input-bordered"
                  value={loginForm.password}
                  onChange={(event) => setLoginForm((prev) => ({ ...prev, password: event.target.value }))}
                />
              </label>
              <button className="btn btn-primary mt-1 w-full" disabled={loading}>
                {loading ? '登录中...' : '登录系统'}
              </button>
            </form>
          ) : (
            <form className="space-y-3" onSubmit={submitRegister}>
              <label className="form-control">
                <span className="label-text mb-1">学生用户名</span>
                <input
                  className="input input-bordered"
                  value={registerForm.username}
                  onChange={(event) => setRegisterForm((prev) => ({ ...prev, username: event.target.value }))}
                />
              </label>
              <label className="form-control">
                <span className="label-text mb-1">密码</span>
                <input
                  type="password"
                  className="input input-bordered"
                  value={registerForm.password}
                  onChange={(event) => setRegisterForm((prev) => ({ ...prev, password: event.target.value }))}
                />
              </label>
              <label className="form-control">
                <span className="label-text mb-1">班级</span>
                <select
                  className="select select-bordered"
                  value={registerForm.classId}
                  onChange={(event) => setRegisterForm((prev) => ({ ...prev, classId: event.target.value }))}
                >
                  <option value="">请选择班级</option>
                  {classOptions.map((item) => (
                    <option key={item.id} value={item.id}>
                      {item.name}
                    </option>
                  ))}
                </select>
              </label>
              <button className="btn btn-secondary mt-1 w-full" disabled={loading}>
                {loading ? '注册中...' : '注册并进入学生端'}
              </button>
            </form>
          )}
        </section>
      </main>
    </div>
  );
}
