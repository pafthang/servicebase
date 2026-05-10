import { CrudService } from "@/services/CrudService";
import { BaseModel, RecordModel } from "@/tools/dtos";
import { CommonOptions } from "@/tools/options";

export interface TeamModel extends BaseModel {
    name: string;
}

export interface TeamMemberModel extends BaseModel {
    team: string;
    userId: string;
    userCollectionId: string;
}

export interface TeamMemberDetails {
    membership: TeamMemberModel;
    user?: RecordModel;
}

export class TeamService extends CrudService<TeamModel> {
    /**
     * @inheritdoc
     */
    get baseCrudPath(): string {
        return "/api/teams";
    }

    /**
     * Returns all teams for the currently authenticated user.
     *
     * @throws {ClientResponseError}
     */
    async getMyTeams(options?: CommonOptions): Promise<Array<TeamModel>> {
        options = Object.assign(
            {
                method: "GET",
            },
            options,
        );

        return this.client.send(this.baseCrudPath + "/me", options);
    }

    /**
     * Returns all members for a team.
     *
     * @throws {ClientResponseError}
     */
    async getMembers(teamId: string, options?: CommonOptions): Promise<Array<TeamMemberDetails>> {
        options = Object.assign(
            {
                method: "GET",
            },
            options,
        );

        return this.client.send(
            this.baseCrudPath + "/" + encodeURIComponent(teamId) + "/members",
            options,
        );
    }

    /**
     * Adds a member to a team.
     *
     * @throws {ClientResponseError}
     */
    async addMember(
        teamId: string,
        userId: string,
        options?: CommonOptions,
    ): Promise<TeamMemberDetails> {
        options = Object.assign(
            {
                method: "POST",
                body: {
                    userId: userId,
                },
            },
            options,
        );

        return this.client.send(
            this.baseCrudPath + "/" + encodeURIComponent(teamId) + "/members",
            options,
        );
    }

    /**
     * Removes a member from a team.
     *
     * @throws {ClientResponseError}
     */
    async removeMember(teamId: string, userId: string, options?: CommonOptions): Promise<boolean> {
        options = Object.assign(
            {
                method: "DELETE",
            },
            options,
        );

        return this.client
            .send(
                this.baseCrudPath +
                    "/" +
                    encodeURIComponent(teamId) +
                    "/members/" +
                    encodeURIComponent(userId),
                options,
            )
            .then(() => true);
    }
}
