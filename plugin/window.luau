-- Palette commands "New Claude chat", "New Codex chat", "New OpenCode chat",
-- "New Pi chat" and "New Hermes chat": run tclaude, tcodex, topencode, tpi or
-- the hermes CLI wrapper, in a pane beside the focused one and in its
-- directory.

local BIN = tern.plugin.dir .. "/../bin/"

local function quote(s: string): string
	return "'" .. string.gsub(s, "'", "'\\''") .. "'"
end

local function open(cx, command: string, extra: string)
	local id = cx.session:focused()
	local pane = if id then cx.session:pane(id) else nil
	local cwd = if pane and pane.cwd ~= "" then pane.cwd else nil
	-- Hermes runs from PATH (its wrapper is installed by hermes-for-tern);
	-- the other agents from bin/ next to this plugin.
	local bin = if command == "hermes" then "" else BIN
	local launch = { command = quote(bin .. command) .. extra, cwd = cwd }
	if id then
		cx.layout:split(id, "right", launch)
	else
		cx.layout:new_tab(launch)
	end
end

local agents = {
	{ id = "claude", name = "Claude", command = "tclaude" },
	{ id = "codex", name = "Codex", command = "tcodex" },
	{ id = "opencode", name = "OpenCode", command = "topencode" },
	{ id = "pi", name = "Pi", command = "tpi" },
	{ id = "hermes", name = "Hermes", command = "hermes" },
}

for _, agent in ipairs(agents) do
	tern.command({
		id = agent.id .. "-chat",
		title = "New " .. agent.name .. " chat",
		icon = "app-" .. agent.id,
		run = function(cx)
			open(cx, agent.command, "")
		end,
	})
	tern.command({
		id = agent.id .. "-continue",
		title = "Continue last " .. agent.name .. " chat",
		icon = "app-" .. agent.id,
		run = function(cx)
			open(cx, agent.command, " -continue")
		end,
	})
end
